package repository

//go:generate go tool mockgen -source=api_key.go -destination=./api_key/api_key_mock.go -package=api_key

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type APIKey interface {
	Create(ctx context.Context, rec entity.APIKeyRecord) (entity.APIKey, error)
	ListByOwner(ctx context.Context, workspaceID, ownerUserID string) ([]entity.APIKey, error)
	ListWorkspaceKeys(ctx context.Context, workspaceID string) ([]entity.APIKey, error)
	GetByID(ctx context.Context, workspaceID, id string) (entity.APIKey, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (entity.APIKey, error)
	Revoke(ctx context.Context, workspaceID, id string) error
	TouchLastUsed(ctx context.Context, id string, at time.Time) error
}
