package audit

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const selectColumns = `id, workspace_id, at, actor_user_id, actor_label, action, target, ip, meta`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Audit {
	return &repo{db: db}
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanEvent(row interface {
	Scan(dest ...any) error
}) (entity.AuditEvent, error) {
	var (
		e           entity.AuditEvent
		workspaceID sql.NullString
		actorUserID sql.NullString
		ip          sql.NullString
		meta        []byte
	)
	if err := row.Scan(&e.ID, &workspaceID, &e.At, &actorUserID, &e.ActorLabel, &e.Action, &e.Target, &ip, &meta); err != nil {
		return entity.AuditEvent{}, err
	}
	e.WorkspaceID = workspaceID.String
	e.ActorUserID = actorUserID.String
	e.IP = ip.String
	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &e.Meta)
	}
	return e, nil
}

func (r *repo) Create(ctx context.Context, e entity.AuditEvent) error {
	var meta any
	if len(e.Meta) > 0 {
		raw, err := json.Marshal(e.Meta)
		if err != nil {
			return fmt.Errorf("marshal audit meta: %w", err)
		}
		meta = raw
	}
	_, err := r.db.Querier(ctx).ExecContext(ctx,
		`INSERT INTO audit_events (workspace_id, actor_user_id, actor_label, action, target, ip, meta)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nullable(e.WorkspaceID), nullable(e.ActorUserID), e.ActorLabel, e.Action, e.Target, nullable(e.IP), meta)
	if err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}
	return nil
}

type cursor struct {
	At time.Time `json:"at"`
	ID string    `json:"id"`
}

func encodeCursor(at time.Time, id string) string {
	raw, _ := json.Marshal(cursor{At: at, ID: id})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeCursor(s string) (cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return cursor{}, err
	}
	var c cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return cursor{}, err
	}
	return c, nil
}

func (r *repo) List(ctx context.Context, workspaceID string, f entity.AuditFilter) ([]entity.AuditEvent, string, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = entity.AuditDefaultLimit
	}
	if limit > entity.AuditMaxLimit {
		limit = entity.AuditMaxLimit
	}

	where := []string{"workspace_id = $1"}
	args := []any{workspaceID}
	placeholder := func(arg any) int {
		args = append(args, arg)
		return len(args)
	}
	if f.ActorUserID != "" {
		where = append(where, fmt.Sprintf("actor_user_id = $%d", placeholder(f.ActorUserID)))
	}
	if f.ActionPrefix != "" {
		where = append(where, fmt.Sprintf("action LIKE $%d", placeholder(f.ActionPrefix+"%")))
	}
	if f.Query != "" {
		p := placeholder("%" + f.Query + "%")
		where = append(where, fmt.Sprintf("(action ILIKE $%d OR target ILIKE $%d)", p, p))
	}
	if f.Cursor != "" {
		c, err := decodeCursor(f.Cursor)
		if err != nil {
			return nil, "", entity.ErrAuditInvalidCursor
		}
		where = append(where, fmt.Sprintf("(at, id) < ($%d, $%d)", placeholder(c.At), placeholder(c.ID)))
	}

	query := `SELECT ` + selectColumns + ` FROM audit_events WHERE ` + strings.Join(where, " AND ") +
		fmt.Sprintf(` ORDER BY at DESC, id DESC LIMIT %d`, limit+1)

	rows, err := r.db.Querier(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list audit events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []entity.AuditEvent
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, "", fmt.Errorf("scan audit event: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate audit events: %w", err)
	}

	next := ""
	if len(out) > limit {
		last := out[limit-1]
		next = encodeCursor(last.At, last.ID)
		out = out[:limit]
	}
	return out, next, nil
}
