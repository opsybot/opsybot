package chat_oauth_state

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

func New(vk valkey.Client) repository.ChatOAuthState {
	return &repo{vk: vk}
}

func key(state string) string { return "chat:oauth:state:" + state }

func (r *repo) Store(ctx context.Context, state string, data entity.ChatOAuthState, ttl time.Duration) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal chat oauth state: %w", err)
	}
	if err := r.vk.Set(ctx, key(state), raw, ttl).Err(); err != nil {
		return fmt.Errorf("store chat oauth state: %w", err)
	}
	return nil
}

func (r *repo) Consume(ctx context.Context, state string) (entity.ChatOAuthState, error) {
	raw, err := r.vk.GetDel(ctx, key(state)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return entity.ChatOAuthState{}, entity.ErrChatOAuthStateInvalid
		}
		return entity.ChatOAuthState{}, fmt.Errorf("consume chat oauth state: %w", err)
	}
	var data entity.ChatOAuthState
	if err := json.Unmarshal(raw, &data); err != nil {
		return entity.ChatOAuthState{}, fmt.Errorf("unmarshal chat oauth state: %w", err)
	}
	return data, nil
}
