package service

//go:generate go tool mockgen -source=users.go -destination=./users/users_mock.go -package=users

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Users interface {
	UpdateProfile(ctx context.Context, in entity.ProfileUpdate) (entity.User, error)
	ChangePassword(ctx context.Context, current, next string) error
	BeginTOTP(ctx context.Context) (entity.TOTPEnrollment, error)
	ConfirmTOTP(ctx context.Context, code string) ([]string, error)
	DisableTOTP(ctx context.Context, code string) error
	RegenerateRecoveryCodes(ctx context.Context, code string) ([]string, error)
	ListSessions(ctx context.Context) ([]entity.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
}
