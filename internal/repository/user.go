package repository

//go:generate go tool mockgen -source=user.go -destination=./user/user_mock.go -package=user

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type User interface {
	Create(ctx context.Context, u entity.NewUser, passwordHash string) (entity.User, error)
	CreateInvited(ctx context.Context, email string) (entity.User, error)
	CreateSSO(ctx context.Context, email, name string) (entity.User, error)
	GetByID(ctx context.Context, id string) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	Activate(ctx context.Context, id, name, passwordHash, timezone string) error
	HasPassword(ctx context.Context, id string) (bool, error)
	PasswordHash(ctx context.Context, id string) (string, error)
	ExistsAny(ctx context.Context) (bool, error)
	TouchLastActive(ctx context.Context, id string) error
	UpdateProfile(ctx context.Context, id string, p entity.ProfileUpdate) error
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	SetTOTP(ctx context.Context, id, secret string) error
	EnableTOTP(ctx context.Context, id string) error
	DisableTOTP(ctx context.Context, id string) error
	TOTPSecret(ctx context.Context, id string) (string, error)
	AcceptTOTPStep(ctx context.Context, id string, step int64) (bool, error)
}
