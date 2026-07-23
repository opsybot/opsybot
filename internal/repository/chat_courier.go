package repository

//go:generate go tool mockgen -source=chat_courier.go -destination=./chat_courier/chat_courier_mock.go -package=chat_courier

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type ChatCourier interface {
	Send(ctx context.Context, in entity.ChatDelivery) (entity.ChatSendResult, error)
	SendToChannel(ctx context.Context, provider entity.ChatProvider, token, guildID, channel, text string) (entity.ChatSendResult, error)
	Validate(ctx context.Context, provider entity.ChatProvider, token, externalID string) (entity.ChatValidation, error)
	LookupUser(ctx context.Context, provider entity.ChatProvider, token, externalID, email string) (entity.ChatUser, error)
	AuthorizeURL(ctx context.Context, provider entity.ChatProvider, scopes []string, redirectURI, state string) (string, error)
	ExchangeOAuth(ctx context.Context, provider entity.ChatProvider, code, redirectURI string) (entity.ChatOAuthResult, error)
	IdentityAuthorizeURL(ctx context.Context, provider entity.ChatProvider, scopes []string, redirectURI, state, teamID string) (string, error)
	ExchangeIdentity(ctx context.Context, provider entity.ChatProvider, code, redirectURI string) (entity.ChatIdentityResult, error)
}
