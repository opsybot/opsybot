package repository

//go:generate go tool mockgen -source=password_reset.go -destination=./password_reset/password_reset_mock.go -package=password_reset

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type PasswordReset interface {
	Create(ctx context.Context, userID, tokenHash, ip string, expiresAt time.Time) error
	GetByTokenHash(ctx context.Context, tokenHash string) (entity.PasswordReset, error)
	MarkUsed(ctx context.Context, id string) error
	DeleteUnusedByUser(ctx context.Context, userID string) error
}
