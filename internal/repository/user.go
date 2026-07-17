package repository

//go:generate go tool mockgen -source=user.go -destination=./user/user_mock.go -package=user

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type User interface {
	Create(ctx context.Context, u entity.NewUser, passwordHash string) (entity.User, error)
	CreateInvited(ctx context.Context, email string) (entity.User, error)
	GetByID(ctx context.Context, id string) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	PasswordHash(ctx context.Context, id string) (string, error)
	ExistsAny(ctx context.Context) (bool, error)
	TouchLastActive(ctx context.Context, id string) error
}
