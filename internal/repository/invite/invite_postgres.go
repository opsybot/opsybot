package invite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const joinedColumns = `i.id, i.workspace_id, i.user_id, i.invited_by, i.status, i.expires_at, i.created_at,
	u.email, inviter.name, w.name, w.slug`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Invite {
	return &repo{db: db}
}

func scanJoined(row interface {
	Scan(dest ...any) error
}) (entity.Invite, error) {
	var (
		inv    entity.Invite
		status string
	)
	err := row.Scan(&inv.ID, &inv.WorkspaceID, &inv.UserID, &inv.InvitedBy, &status, &inv.ExpiresAt, &inv.CreatedAt,
		&inv.Email, &inv.InvitedByName, &inv.WorkspaceName, &inv.WorkspaceSlug)
	inv.Status = entity.InviteStatus(status)
	return inv, err
}

func (r *repo) get(ctx context.Context, where string, args ...any) (entity.Invite, error) {
	inv, err := scanJoined(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+joinedColumns+`
		 FROM invites i
		 JOIN users u ON u.id = i.user_id
		 JOIN users inviter ON inviter.id = i.invited_by
		 JOIN workspaces w ON w.id = i.workspace_id
		 WHERE `+where, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Invite{}, entity.ErrInviteNotFound
		}
		return entity.Invite{}, fmt.Errorf("get invite: %w", err)
	}
	return inv, nil
}

func (r *repo) Create(ctx context.Context, workspaceID, userID, invitedBy, tokenHash string, expiresAt time.Time) (entity.Invite, error) {
	m := &dbpostgres.Invite{
		WorkspaceID: workspaceID,
		UserID:      userID,
		InvitedBy:   invitedBy,
		TokenHash:   tokenHash,
		ExpiresAt:   expiresAt,
	}
	if err := m.Insert(ctx, r.db.Querier(ctx),
		boil.Whitelist("workspace_id", "user_id", "invited_by", "token_hash", "expires_at")); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.Invite{}, entity.ErrInvitePending
		}
		return entity.Invite{}, fmt.Errorf("create invite: %w", err)
	}
	return r.get(ctx, "i.id = $1", m.ID)
}

func (r *repo) GetByTokenHash(ctx context.Context, tokenHash string) (entity.Invite, error) {
	return r.get(ctx, "i.token_hash = $1", tokenHash)
}

func (r *repo) GetPending(ctx context.Context, workspaceID, userID string) (entity.Invite, error) {
	return r.get(ctx, "i.workspace_id = $1 AND i.user_id = $2 AND i.status = 'pending'", workspaceID, userID)
}

func (r *repo) ListPending(ctx context.Context, workspaceID string) ([]entity.Invite, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT `+joinedColumns+`
		 FROM invites i
		 JOIN users u ON u.id = i.user_id
		 JOIN users inviter ON inviter.id = i.invited_by
		 JOIN workspaces w ON w.id = i.workspace_id
		 WHERE i.workspace_id = $1 AND i.status = 'pending'
		 ORDER BY i.created_at DESC`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list pending invites: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []entity.Invite
	for rows.Next() {
		inv, err := scanJoined(rows)
		if err != nil {
			return nil, fmt.Errorf("scan invite: %w", err)
		}
		out = append(out, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invites: %w", err)
	}
	return out, nil
}

func (r *repo) RotateToken(ctx context.Context, id, tokenHash string, expiresAt time.Time) error {
	if _, err := dbpostgres.Invites(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"token_hash": tokenHash, "expires_at": expiresAt, "updated_at": time.Now()}); err != nil {
		return fmt.Errorf("rotate invite token: %w", err)
	}
	return nil
}

func (r *repo) MarkAccepted(ctx context.Context, id string) error {
	now := time.Now()
	if _, err := dbpostgres.Invites(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"status": "accepted", "accepted_at": now, "updated_at": now}); err != nil {
		return fmt.Errorf("mark invite accepted: %w", err)
	}
	return nil
}

func (r *repo) MarkRevoked(ctx context.Context, id string) error {
	if _, err := dbpostgres.Invites(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"status": "revoked", "updated_at": time.Now()}); err != nil {
		return fmt.Errorf("mark invite revoked: %w", err)
	}
	return nil
}
