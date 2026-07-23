package service

//go:generate go tool mockgen -source=interactions.go -destination=./interactions/interactions_mock.go -package=interactions

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Interactions interface {
	Slack(ctx context.Context, cb entity.ChatCallback) (entity.InteractionResponse, error)
	Discord(ctx context.Context, cb entity.ChatCallback) (entity.InteractionResponse, error)
}
