package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const selectColumns = `id, user_id, ip, user_agent, expires_at, last_seen_at, created_at`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Session {
	return &repo{db: db}
}

func scanSession(row interface {
	Scan(dest ...any) error
}) (entity.Session, error) {
	var (
		m  dbpostgres.Session
		ip sql.NullString
	)
	if err := row.Scan(&m.ID, &m.UserID, &ip, &m.UserAgent, &m.ExpiresAt, &m.LastSeenAt, &m.CreatedAt); err != nil {
		return entity.Session{}, err
	}
	return entity.Session{
		ID:         m.ID,
		UserID:     m.UserID,
		IP:         ip.String,
		UserAgent:  m.UserAgent,
		ExpiresAt:  m.ExpiresAt,
		LastSeenAt: m.LastSeenAt,
		CreatedAt:  m.CreatedAt,
	}, nil
}

func nullIP(ip string) any {
	if ip == "" {
		return nil
	}
	return ip
}

func (r *repo) Create(ctx context.Context, userID, tokenHash, ip, userAgent string, expiresAt time.Time) (entity.Session, error) {
	m, err := scanSession(r.db.Querier(ctx).QueryRowContext(ctx,
		`INSERT INTO sessions (user_id, token_hash, ip, user_agent, expires_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING `+selectColumns,
		userID, tokenHash, nullIP(ip), userAgent, expiresAt))
	if err != nil {
		return entity.Session{}, fmt.Errorf("create session: %w", err)
	}
	return m, nil
}

func (r *repo) GetByTokenHash(ctx context.Context, tokenHash string) (entity.Session, error) {
	m, err := scanSession(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+selectColumns+` FROM sessions
		 WHERE token_hash = $1 AND revoked_at IS NULL`, tokenHash))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Session{}, entity.ErrSessionNotFound
		}
		return entity.Session{}, fmt.Errorf("get session: %w", err)
	}
	return m, nil
}

func (r *repo) Touch(ctx context.Context, id string, seenAt, expiresAt time.Time) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE sessions SET last_seen_at = $2, expires_at = $3 WHERE id = $1`,
		id, seenAt, expiresAt); err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *repo) DeleteByUser(ctx context.Context, userID string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
		return fmt.Errorf("delete sessions by user: %w", err)
	}
	return nil
}
