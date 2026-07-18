package password_reset

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

func New(db postgres.Client) repository.PasswordReset {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, userID, tokenHash, ip string, expiresAt time.Time) error {
	m := &dbpostgres.PasswordResetToken{
		UserID:    userID,
		TokenHash: tokenHash,
		RequestIP: null.NewString(ip, ip != ""),
		ExpiresAt: expiresAt,
	}
	if err := m.Insert(ctx, r.db.Querier(ctx),
		boil.Whitelist("user_id", "token_hash", "request_ip", "expires_at")); err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}
	return nil
}

func (r *repo) GetByTokenHash(ctx context.Context, tokenHash string) (entity.PasswordReset, error) {
	m, err := dbpostgres.PasswordResetTokens(qm.Where("token_hash = ?", tokenHash)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.PasswordReset{}, entity.ErrPasswordResetNotFound
		}
		return entity.PasswordReset{}, fmt.Errorf("get password reset: %w", err)
	}
	return entity.PasswordReset{
		ID:        m.ID,
		UserID:    m.UserID,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt.Time,
		CreatedAt: m.CreatedAt,
	}, nil
}

func (r *repo) MarkUsed(ctx context.Context, id string) error {
	if _, err := dbpostgres.PasswordResetTokens(qm.Where("id = ?", id)).
		UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"used_at": time.Now()}); err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}
	return nil
}

func (r *repo) DeleteUnusedByUser(ctx context.Context, userID string) error {
	if _, err := dbpostgres.PasswordResetTokens(
		qm.Where("user_id = ?", userID),
		qm.Where("used_at IS NULL"),
	).UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"used_at": time.Now()}); err != nil {
		return fmt.Errorf("delete unused password resets: %w", err)
	}
	return nil
}
