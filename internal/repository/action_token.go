package repository

//go:generate go tool mockgen -source=action_token.go -destination=./action_token/action_token_mock.go -package=action_token

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type ActionToken interface {
	Mint(ctx context.Context, rec entity.AlertActionRecord) error
	Consume(ctx context.Context, tokenHash, ip string, now time.Time) (entity.ActionClaim, error)
	DeleteForAlert(ctx context.Context, alertID string) error
}
