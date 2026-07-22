package repository

//go:generate go tool mockgen -source=channel.go -destination=./channel/channel_mock.go -package=channel

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Channel interface {
	Create(ctx context.Context, userID string, c entity.NewChannel) (entity.Channel, error)
	ListByUser(ctx context.Context, userID string) ([]entity.Channel, error)
	ListByUsers(ctx context.Context, userIDs []string) (map[string][]entity.Channel, error)
	Get(ctx context.Context, id, userID string) (entity.Channel, error)
	MarkVerified(ctx context.Context, id, userID string) error
	Delete(ctx context.Context, id, userID string) error
}
