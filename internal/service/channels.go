package service

//go:generate go tool mockgen -source=channels.go -destination=./channels/channels_mock.go -package=channels

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Channels interface {
	List(ctx context.Context) ([]entity.Channel, error)
	Add(ctx context.Context, in entity.NewChannel) (entity.Channel, error)
	Verify(ctx context.Context, channelID string) error
	Remove(ctx context.Context, channelID string) error
}
