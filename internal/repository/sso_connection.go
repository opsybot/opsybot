package repository

//go:generate go tool mockgen -source=sso_connection.go -destination=./sso_connection/sso_connection_mock.go -package=sso_connection

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type SSOConnection interface {
	Get(ctx context.Context, workspaceID string) (entity.SSOConnection, error)
	Save(ctx context.Context, workspaceID string, in entity.SSOConfigInput) (entity.SSOConnection, error)
	ClientSecret(ctx context.Context, workspaceID string) (string, error)
}
