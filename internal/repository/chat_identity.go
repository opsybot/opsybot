package repository

//go:generate go tool mockgen -source=chat_identity.go -destination=./chat_identity/chat_identity_mock.go -package=chat_identity

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type ChatIdentity interface {
	Upsert(ctx context.Context, in entity.ChatIdentity) (entity.ChatIdentity, error)
	GetForUser(ctx context.Context, connectionID, userID string) (entity.ChatIdentity, error)
	SetDMChannel(ctx context.Context, id, dmChannelID string) error
}
