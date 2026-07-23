package repository

//go:generate go tool mockgen -source=chat_oauth_state.go -destination=./chat_oauth_state/chat_oauth_state_mock.go -package=chat_oauth_state

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type ChatOAuthState interface {
	Store(ctx context.Context, state string, data entity.ChatOAuthState, ttl time.Duration) error
	Consume(ctx context.Context, state string) (entity.ChatOAuthState, error)
}
