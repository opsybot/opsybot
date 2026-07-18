package repository

//go:generate go tool mockgen -source=ratelimit.go -destination=./ratelimit/ratelimit_mock.go -package=ratelimit

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit entity.RateLimit) (entity.RateResult, error)
}
