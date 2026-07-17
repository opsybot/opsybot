package service

//go:generate go tool mockgen -source=apikeys.go -destination=./apikeys/apikeys_mock.go -package=apikeys

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type APIKeys interface {
	List(ctx context.Context, workspaceSlug string) (entity.APIKeyList, error)
	Create(ctx context.Context, workspaceSlug string, in entity.NewAPIKey) (entity.APIKey, string, error)
	Revoke(ctx context.Context, workspaceSlug, keyID string) error
	Resolve(ctx context.Context, secret string) (entity.Identity, error)
}
