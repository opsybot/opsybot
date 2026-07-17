package service

//go:generate go tool mockgen -source=sso.go -destination=./sso/sso_mock.go -package=sso

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type SSO interface {
	GetConfig(ctx context.Context, workspaceSlug string) (entity.SSOConnection, error)
	SaveConfig(ctx context.Context, workspaceSlug string, in entity.SSOConfigInput) (entity.SSOConnection, error)
	StartLogin(ctx context.Context, workspaceSlug string) (string, error)
	CompleteLogin(ctx context.Context, workspaceSlug, code, state, ip, userAgent string) (entity.LoginResult, error)
}
