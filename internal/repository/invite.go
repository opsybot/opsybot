package repository

//go:generate go tool mockgen -source=invite.go -destination=./invite/invite_mock.go -package=invite

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Invite interface {
	Create(ctx context.Context, workspaceID, userID, invitedBy, tokenHash string, expiresAt time.Time) (entity.Invite, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (entity.Invite, error)
	GetPending(ctx context.Context, workspaceID, userID string) (entity.Invite, error)
	ListPending(ctx context.Context, workspaceID string) ([]entity.Invite, error)
	RotateToken(ctx context.Context, id, tokenHash string, expiresAt time.Time) error
	MarkAccepted(ctx context.Context, id string) error
	MarkRevoked(ctx context.Context, id string) error
}
