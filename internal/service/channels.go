package service

//go:generate go tool mockgen -source=channels.go -destination=./channels/channels_mock.go -package=channels

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Channels interface {
	List(ctx context.Context) ([]entity.Channel, error)
	Add(ctx context.Context, in entity.NewChannel) (entity.Channel, error)
	StartVerification(ctx context.Context, channelID string) (entity.ChannelVerification, error)
	CompleteVerification(ctx context.Context, channelID, code string) error
	CompleteByToken(ctx context.Context, token string) error
	SendTest(ctx context.Context, channelID string) (entity.NotifyResult, error)
	Remove(ctx context.Context, channelID string) error
}
