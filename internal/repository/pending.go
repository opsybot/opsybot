package repository

//go:generate go tool mockgen -source=pending.go -destination=./pending/pending_mock.go -package=pending

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Pending interface {
	Store(ctx context.Context, tokenHash string, p entity.PendingTwoFactor, ttl time.Duration) error
	Get(ctx context.Context, tokenHash string) (entity.PendingTwoFactor, error)
	Delete(ctx context.Context, tokenHash string) error
	IncrAttempts(ctx context.Context, tokenHash string) (int, error)
}
