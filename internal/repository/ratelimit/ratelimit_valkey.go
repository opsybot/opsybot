package ratelimit

import (
	"context"
	"fmt"

	"github.com/go-redis/redis_rate/v10"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/valkey"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	limiter *redis_rate.Limiter
}

func New(vk valkey.Client) repository.RateLimiter {
	return &repo{limiter: redis_rate.NewLimiter(vk)}
}

func (r *repo) Allow(ctx context.Context, key string, limit entity.RateLimit) (entity.RateResult, error) {
	res, err := r.limiter.Allow(ctx, key, redis_rate.Limit{Rate: limit.Rate, Period: limit.Period, Burst: limit.Burst})
	if err != nil {
		return entity.RateResult{}, fmt.Errorf("rate limit allow: %w", err)
	}
	return entity.RateResult{Allowed: res.Allowed > 0, RetryAfter: res.RetryAfter}, nil
}
