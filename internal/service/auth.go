package service

//go:generate go tool mockgen -source=auth.go -destination=./auth/auth_mock.go -package=auth

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Auth interface {
	SetupRequired(ctx context.Context) (bool, error)
	Setup(ctx context.Context, in entity.Setup, ip, userAgent string) (entity.SetupResult, error)
	Login(ctx context.Context, in entity.LoginInput) (entity.LoginResult, error)
	Logout(ctx context.Context) error
	Resolve(ctx context.Context, token string) (entity.Identity, error)
	Profile(ctx context.Context) (entity.User, error)
}
