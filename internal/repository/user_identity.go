package repository

//go:generate go tool mockgen -source=user_identity.go -destination=./user_identity/user_identity_mock.go -package=user_identity

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type UserIdentity interface {
	GetBySubject(ctx context.Context, connectionID, subject string) (entity.UserIdentity, error)
	Create(ctx context.Context, userID, connectionID, subject, email string) error
}
