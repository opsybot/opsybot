package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const selectColumns = `m.workspace_id, m.user_id, m.status, m.joined_at, m.deactivated_at,
	u.name, u.email, u.password_hash, u.totp_enabled_at, u.last_active_at,
	EXISTS (SELECT 1 FROM user_identities ui JOIN sso_connections c ON c.id = ui.connection_id
		WHERE ui.user_id = m.user_id AND c.workspace_id = m.workspace_id)`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Member {
	return &repo{db: db}
}

func scanMember(row interface {
	Scan(dest ...any) error
}) (entity.Member, error) {
	var (
		m             entity.Member
		status        string
		joinedAt      sql.NullTime
		deactivatedAt sql.NullTime
		passwordHash  sql.NullString
		totpEnabledAt sql.NullTime
		lastActiveAt  sql.NullTime
		hasSSO        bool
	)
	if err := row.Scan(&m.WorkspaceID, &m.UserID, &status, &joinedAt, &deactivatedAt,
		&m.Name, &m.Email, &passwordHash, &totpEnabledAt, &lastActiveAt, &hasSSO); err != nil {
		return entity.Member{}, err
	}
	m.Status = entity.MemberStatus(status)
	m.JoinedAt = joinedAt.Time
	m.DeactivatedAt = deactivatedAt.Time
	m.HasPassword = passwordHash.Valid
	m.TOTPEnabled = totpEnabledAt.Valid
	m.LastActiveAt = lastActiveAt.Time
	m.HasSSO = hasSSO
	return m, nil
}

func (r *repo) Create(ctx context.Context, workspaceID, userID string, status entity.MemberStatus) error {
	joined := "NULL"
	if status != entity.MemberStatusInvited {
		joined = "now()"
	}
	_, err := r.db.Querier(ctx).ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, status, joined_at) VALUES ($1, $2, $3, `+joined+`)`,
		workspaceID, userID, string(status))
	if err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.ErrMemberAlreadyExists
		}
		return fmt.Errorf("create member: %w", err)
	}
	return nil
}

func (r *repo) Get(ctx context.Context, workspaceID, userID string) (entity.Member, error) {
	m, err := scanMember(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+selectColumns+`
		 FROM workspace_members m JOIN users u ON u.id = m.user_id
		 WHERE m.workspace_id = $1 AND m.user_id = $2`, workspaceID, userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Member{}, entity.ErrMemberNotFound
		}
		return entity.Member{}, fmt.Errorf("get member: %w", err)
	}
	return m, nil
}

func (r *repo) ListByWorkspace(ctx context.Context, workspaceID string) ([]entity.Member, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT `+selectColumns+`
		 FROM workspace_members m JOIN users u ON u.id = m.user_id
		 WHERE m.workspace_id = $1 ORDER BY u.name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []entity.Member
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}
	return out, nil
}

func (r *repo) IsActive(ctx context.Context, workspaceID, userID string) (bool, error) {
	var active bool
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND status = 'active')`,
		workspaceID, userID).Scan(&active)
	if err != nil {
		return false, fmt.Errorf("member is active: %w", err)
	}
	return active, nil
}

func (r *repo) UpdateStatus(ctx context.Context, workspaceID, userID string, status entity.MemberStatus) error {
	joined := ""
	deactivated := "deactivated_at = NULL"
	switch status {
	case entity.MemberStatusActive:
		joined = "joined_at = COALESCE(joined_at, now()), "
	case entity.MemberStatusDeactivated:
		deactivated = "deactivated_at = now()"
	}
	res, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE workspace_members SET status = $3, `+joined+deactivated+`, updated_at = now()
		 WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID, string(status))
	if err != nil {
		return fmt.Errorf("update member status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update member status rows: %w", err)
	}
	if n == 0 {
		return entity.ErrMemberNotFound
	}
	return nil
}

func (r *repo) DeleteInvited(ctx context.Context, workspaceID, userID string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND status = 'invited'`,
		workspaceID, userID); err != nil {
		return fmt.Errorf("delete invited member: %w", err)
	}
	return nil
}

func (r *repo) CountOtherActive(ctx context.Context, userID, exceptWorkspaceID string) (int, error) {
	var count int
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT count(*) FROM workspace_members WHERE user_id = $1 AND workspace_id <> $2 AND status = 'active'`,
		userID, exceptWorkspaceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count other active memberships: %w", err)
	}
	return count, nil
}
