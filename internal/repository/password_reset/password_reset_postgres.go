package password_reset

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	var reqIP any
	if ip != "" {
		reqIP = ip
	}
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, request_ip, expires_at) VALUES ($1, $2, $3, $4)`,
		userID, tokenHash, reqIP, expiresAt); err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}
	return nil
}

func (r *repo) GetByTokenHash(ctx context.Context, tokenHash string) (entity.PasswordReset, error) {
	var (
		p      entity.PasswordReset
		usedAt sql.NullTime
	)
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT id, user_id, expires_at, used_at, created_at FROM password_reset_tokens WHERE token_hash = $1`,
		tokenHash).Scan(&p.ID, &p.UserID, &p.ExpiresAt, &usedAt, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.PasswordReset{}, entity.ErrPasswordResetNotFound
		}
		return entity.PasswordReset{}, fmt.Errorf("get password reset: %w", err)
	}
	p.UsedAt = usedAt.Time
	return p, nil
}

func (r *repo) MarkUsed(ctx context.Context, id string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = now() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}
	return nil
}

func (r *repo) DeleteUnusedByUser(ctx context.Context, userID string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = now() WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
		return fmt.Errorf("delete unused password resets: %w", err)
	}
	return nil
}
