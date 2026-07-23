package action_token

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.ActionToken {
	return &repo{db: db}
}

const consumeSQL = `
UPDATE alert_action_tokens
   SET used_at = $2, used_ip = NULLIF($3, '')::inet
 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2
RETURNING workspace_id, alert_id, user_id, action`

func (r *repo) Mint(ctx context.Context, rec entity.AlertActionRecord) error {
	m := &dbpostgres.AlertActionToken{
		WorkspaceID: rec.WorkspaceID,
		AlertID:     rec.AlertID,
		UserID:      rec.UserID,
		Action:      string(rec.Action),
		TokenHash:   rec.TokenHash,
		ExpiresAt:   rec.ExpiresAt,
	}
	if rec.ChannelID != "" {
		m.ChannelID = null.StringFrom(rec.ChannelID)
	}
	cols := boil.Whitelist("workspace_id", "alert_id", "user_id", "channel_id", "action", "token_hash", "expires_at")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		return fmt.Errorf("mint action token: %w", err)
	}
	return nil
}

func (r *repo) Consume(ctx context.Context, tokenHash, ip string, now time.Time) (entity.ActionClaim, error) {
	var claim entity.ActionClaim
	var action string
	err := r.db.Querier(ctx).QueryRowContext(ctx, consumeSQL, tokenHash, now, ip).
		Scan(&claim.WorkspaceID, &claim.AlertID, &claim.UserID, &action)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ActionClaim{}, entity.ErrActionTokenInvalid
		}
		return entity.ActionClaim{}, fmt.Errorf("consume action token: %w", err)
	}
	claim.Action = entity.ActionKind(action)
	return claim, nil
}

func (r *repo) DeleteForAlert(ctx context.Context, alertID string) error {
	_, err := dbpostgres.AlertActionTokens(qm.Where("alert_id = ? AND used_at IS NULL", alertID)).
		DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("delete action tokens: %w", err)
	}
	return nil
}
