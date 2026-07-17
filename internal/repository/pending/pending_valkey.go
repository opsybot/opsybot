package pending

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/valkey"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	vk valkey.Client
}

func New(vk valkey.Client) repository.Pending {
	return &repo{vk: vk}
}

func key(tokenHash string) string         { return "2fa:" + tokenHash }
func attemptsKey(tokenHash string) string { return "2fa:attempts:" + tokenHash }

func (r *repo) Store(ctx context.Context, tokenHash string, p entity.PendingTwoFactor, ttl time.Duration) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal pending: %w", err)
	}
	if err := r.vk.Set(ctx, key(tokenHash), raw, ttl).Err(); err != nil {
		return fmt.Errorf("store pending: %w", err)
	}
	return nil
}

func (r *repo) Get(ctx context.Context, tokenHash string) (entity.PendingTwoFactor, error) {
	raw, err := r.vk.Get(ctx, key(tokenHash)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return entity.PendingTwoFactor{}, entity.ErrPendingNotFound
		}
		return entity.PendingTwoFactor{}, fmt.Errorf("get pending: %w", err)
	}
	var p entity.PendingTwoFactor
	if err := json.Unmarshal(raw, &p); err != nil {
		return entity.PendingTwoFactor{}, fmt.Errorf("unmarshal pending: %w", err)
	}
	return p, nil
}

func (r *repo) Delete(ctx context.Context, tokenHash string) error {
	if err := r.vk.Del(ctx, key(tokenHash), attemptsKey(tokenHash)).Err(); err != nil {
		return fmt.Errorf("delete pending: %w", err)
	}
	return nil
}

func (r *repo) IncrAttempts(ctx context.Context, tokenHash string) (int, error) {
	n, err := r.vk.Incr(ctx, attemptsKey(tokenHash)).Result()
	if err != nil {
		return 0, fmt.Errorf("incr pending attempts: %w", err)
	}
	if n == 1 {
		_ = r.vk.Expire(ctx, attemptsKey(tokenHash), entity.PendingTwoFactorTTL).Err()
	}
	return int(n), nil
}
