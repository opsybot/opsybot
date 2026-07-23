package service

//go:generate go tool mockgen -source=chats.go -destination=./chats/chats_mock.go -package=chats

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Chats interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.ChatConnection, error)
	Connect(ctx context.Context, workspaceSlug string, in entity.ChatConnectInput) (entity.ChatConnection, error)
	Delete(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) error
	SetDefaults(ctx context.Context, workspaceSlug string, provider entity.ChatProvider, namingPattern, announceChannel string, archiveOnResolve bool) error
	LinkIdentity(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) (entity.ChatIdentity, error)
	TestConnection(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) (entity.ChatSendResult, error)
	StartOAuth(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) (string, error)
	CompleteOAuth(ctx context.Context, provider entity.ChatProvider, code, guildID, state string) (string, error)
	StartIdentityOAuth(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) (string, error)
	CompleteIdentityOAuth(ctx context.Context, provider entity.ChatProvider, code, state string) (string, error)
	StartTelegramLink(ctx context.Context, workspaceSlug string) (string, error)
	CompleteTelegramLink(ctx context.Context, token, telegramUserID, handle string) error
	AnswerTelegramCallback(ctx context.Context, callbackID, text string) error
}
