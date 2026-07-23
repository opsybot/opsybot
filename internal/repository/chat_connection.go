package repository

//go:generate go tool mockgen -source=chat_connection.go -destination=./chat_connection/chat_connection_mock.go -package=chat_connection

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type ChatConnection interface {
	List(ctx context.Context, workspaceID string) ([]entity.ChatConnection, error)
	Get(ctx context.Context, workspaceID string, provider entity.ChatProvider) (entity.ChatConnection, error)
	BotToken(ctx context.Context, workspaceID string, provider entity.ChatProvider) (string, error)
	Save(ctx context.Context, workspaceID string, in entity.ChatConnectionInput) (entity.ChatConnection, error)
	SecretsEnabled(ctx context.Context) bool
	SetHealth(ctx context.Context, workspaceID string, provider entity.ChatProvider, health entity.ChatHealth, note string, at time.Time) error
	SetDefaults(ctx context.Context, workspaceID string, provider entity.ChatProvider, namingPattern, announceChannel string, archiveOnResolve bool) error
	Delete(ctx context.Context, workspaceID string, provider entity.ChatProvider) error
}
