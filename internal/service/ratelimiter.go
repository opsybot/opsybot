package service

//go:generate go tool mockgen -source=ratelimiter.go -destination=./ratelimiter/ratelimiter_mock.go -package=ratelimiter

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type RateLimiter interface {
	Allow(ctx context.Context, scope entity.RateScope, key string) (entity.RateResult, error)
}
