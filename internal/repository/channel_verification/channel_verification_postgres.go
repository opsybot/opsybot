package channel_verification

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

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.ChannelVerification {
	return &repo{db: db}
}

const consumeTokenSQL = `
UPDATE channel_verifications
   SET used_at = $2
 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2
RETURNING channel_id, user_id`

const consumeCodeSQL = `
UPDATE channel_verifications
   SET used_at = $4
 WHERE id = (
   SELECT id FROM channel_verifications
    WHERE channel_id = $1 AND user_id = $2 AND used_at IS NULL AND expires_at > $4
      AND code_hash <> '' AND code_hash = $3
    ORDER BY created_at DESC
    LIMIT 1
 )
RETURNING channel_id, user_id`

func (r *repo) Start(ctx context.Context, rec entity.ChannelVerifyRecord) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.ChannelVerifications(
		qm.Where("channel_id = ? AND used_at IS NULL", rec.ChannelID),
	).UpdateAll(ctx, exec, dbpostgres.M{"used_at": time.Now().UTC()}); err != nil {
		return fmt.Errorf("invalidate prior verifications: %w", err)
	}
	m := &dbpostgres.ChannelVerification{
		ChannelID: rec.ChannelID,
		UserID:    rec.UserID,
		Method:    string(rec.Method),
		TokenHash: rec.TokenHash,
		CodeHash:  rec.CodeHash,
		Nonce:     rec.Nonce,
		ExpiresAt: rec.ExpiresAt,
	}
	cols := boil.Whitelist("channel_id", "user_id", "method", "token_hash", "code_hash", "nonce", "expires_at")
	if err := m.Insert(ctx, exec, cols); err != nil {
		return fmt.Errorf("start channel verification: %w", err)
	}
	return nil
}

func (r *repo) ConsumeToken(ctx context.Context, tokenHash string, now time.Time) (entity.ChannelVerifyClaim, error) {
	var claim entity.ChannelVerifyClaim
	err := r.db.Querier(ctx).QueryRowContext(ctx, consumeTokenSQL, tokenHash, now).Scan(&claim.ChannelID, &claim.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ChannelVerifyClaim{}, entity.ErrChannelVerifyInvalid
		}
		return entity.ChannelVerifyClaim{}, fmt.Errorf("consume verification token: %w", err)
	}
	return claim, nil
}

func (r *repo) ConsumeCode(ctx context.Context, channelID, userID, codeHash string, now time.Time) (entity.ChannelVerifyClaim, error) {
	var claim entity.ChannelVerifyClaim
	err := r.db.Querier(ctx).QueryRowContext(ctx, consumeCodeSQL, channelID, userID, codeHash, now).Scan(&claim.ChannelID, &claim.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ChannelVerifyClaim{}, entity.ErrChannelVerifyInvalid
		}
		return entity.ChannelVerifyClaim{}, fmt.Errorf("consume verification code: %w", err)
	}
	return claim, nil
}
