package session

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

func New(db postgres.Client) repository.Session {
	return &repo{db: db}
}

func toEntity(m *dbpostgres.Session) entity.Session {
	return entity.Session{
		ID:                m.ID,
		UserID:            m.UserID,
		IP:                m.IP.String,
		UserAgent:         m.UserAgent,
		ExpiresAt:         m.ExpiresAt,
		AbsoluteExpiresAt: m.AbsoluteExpiresAt.Time,
		LastSeenAt:        m.LastSeenAt,
		CreatedAt:         m.CreatedAt,
	}
}

func (r *repo) Create(ctx context.Context, userID, tokenHash, ip, userAgent string, expiresAt, absoluteExpiresAt time.Time) (entity.Session, error) {
	m := &dbpostgres.Session{
		UserID:            userID,
		TokenHash:         tokenHash,
		IP:                null.NewString(ip, ip != ""),
		UserAgent:         userAgent,
		ExpiresAt:         expiresAt,
		AbsoluteExpiresAt: null.TimeFrom(absoluteExpiresAt),
	}
	if err := m.Insert(ctx, r.db.Querier(ctx),
		boil.Whitelist("user_id", "token_hash", "ip", "user_agent", "expires_at", "absolute_expires_at")); err != nil {
		return entity.Session{}, fmt.Errorf("create session: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetByTokenHash(ctx context.Context, tokenHash string) (entity.Session, error) {
	m, err := dbpostgres.Sessions(
		qm.Where("token_hash = ?", tokenHash),
		qm.Where("revoked_at IS NULL"),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Session{}, entity.ErrSessionNotFound
		}
		return entity.Session{}, fmt.Errorf("get session: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) Touch(ctx context.Context, id string, seenAt, expiresAt time.Time) error {
	if _, err := dbpostgres.Sessions(qm.Where("id = ?", id)).
		UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"last_seen_at": seenAt, "expires_at": expiresAt}); err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	if _, err := dbpostgres.Sessions(qm.Where("id = ?", id), qm.Where("revoked_at IS NULL")).
		UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"revoked_at": time.Now()}); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *repo) DeleteByUser(ctx context.Context, userID string) error {
	if _, err := dbpostgres.Sessions(qm.Where("user_id = ?", userID), qm.Where("revoked_at IS NULL")).
		UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"revoked_at": time.Now()}); err != nil {
		return fmt.Errorf("delete sessions by user: %w", err)
	}
	return nil
}

func (r *repo) DeleteOthers(ctx context.Context, userID, exceptSessionID string) error {
	if _, err := dbpostgres.Sessions(
		qm.Where("user_id = ?", userID),
		qm.Where("id <> ?", exceptSessionID),
		qm.Where("revoked_at IS NULL"),
	).UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"revoked_at": time.Now()}); err != nil {
		return fmt.Errorf("delete other sessions: %w", err)
	}
	return nil
}

func (r *repo) OwnedBy(ctx context.Context, id, userID string) (bool, error) {
	owned, err := dbpostgres.Sessions(
		qm.Where("id = ? AND user_id = ?", id, userID),
		qm.Where("revoked_at IS NULL"),
	).Exists(ctx, r.db.Querier(ctx))
	if err != nil {
		return false, fmt.Errorf("session owned by: %w", err)
	}
	return owned, nil
}

func (r *repo) ListByUser(ctx context.Context, userID string) ([]entity.Session, error) {
	rows, err := dbpostgres.Sessions(
		qm.Where("user_id = ?", userID),
		qm.Where("revoked_at IS NULL"),
		qm.Where("expires_at > now()"),
		qm.OrderBy("last_seen_at DESC"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	out := make([]entity.Session, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEntity(m))
	}
	return out, nil
}
