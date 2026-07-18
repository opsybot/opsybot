package repository

//go:generate go tool mockgen -source=session.go -destination=./session/session_mock.go -package=session

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Session interface {
	Create(ctx context.Context, userID, tokenHash, ip, userAgent string, expiresAt, absoluteExpiresAt time.Time) (entity.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (entity.Session, error)
	Touch(ctx context.Context, id string, seenAt, expiresAt time.Time) error
	Delete(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, userID string) error
	ListByUser(ctx context.Context, userID string) ([]entity.Session, error)
	DeleteOthers(ctx context.Context, userID, exceptSessionID string) error
	OwnedBy(ctx context.Context, id, userID string) (bool, error)
}
