package repository

//go:generate go tool mockgen -source=sso_state.go -destination=./sso_state/sso_state_mock.go -package=sso_state

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type SSOState interface {
	Store(ctx context.Context, state string, data entity.SSOState, ttl time.Duration) error
	Consume(ctx context.Context, state string) (entity.SSOState, error)
}
