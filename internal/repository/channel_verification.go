package repository

//go:generate go tool mockgen -source=channel_verification.go -destination=./channel_verification/channel_verification_mock.go -package=channel_verification

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type ChannelVerification interface {
	Start(ctx context.Context, rec entity.ChannelVerifyRecord) error
	ConsumeToken(ctx context.Context, tokenHash string, now time.Time) (entity.ChannelVerifyClaim, error)
	ConsumeCode(ctx context.Context, channelID, userID, codeHash string, now time.Time) (entity.ChannelVerifyClaim, error)
}
