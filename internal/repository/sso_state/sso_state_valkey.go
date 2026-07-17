package sso_state

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

func New(vk valkey.Client) repository.SSOState {
	return &repo{vk: vk}
}

func key(state string) string { return "sso:state:" + state }

func (r *repo) Store(ctx context.Context, state string, data entity.SSOState, ttl time.Duration) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal sso state: %w", err)
	}
	if err := r.vk.Set(ctx, key(state), raw, ttl).Err(); err != nil {
		return fmt.Errorf("store sso state: %w", err)
	}
	return nil
}

func (r *repo) Consume(ctx context.Context, state string) (entity.SSOState, error) {
	raw, err := r.vk.GetDel(ctx, key(state)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return entity.SSOState{}, entity.ErrSSOStateInvalid
		}
		return entity.SSOState{}, fmt.Errorf("consume sso state: %w", err)
	}
	var data entity.SSOState
	if err := json.Unmarshal(raw, &data); err != nil {
		return entity.SSOState{}, fmt.Errorf("unmarshal sso state: %w", err)
	}
	return data, nil
}
