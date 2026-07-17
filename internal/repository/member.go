package repository

//go:generate go tool mockgen -source=member.go -destination=./member/member_mock.go -package=member

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Member interface {
	Create(ctx context.Context, workspaceID, userID string, status entity.MemberStatus) error
	Get(ctx context.Context, workspaceID, userID string) (entity.Member, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]entity.Member, error)
	IsActive(ctx context.Context, workspaceID, userID string) (bool, error)
}
