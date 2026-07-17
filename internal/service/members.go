package service

//go:generate go tool mockgen -source=members.go -destination=./members/members_mock.go -package=members

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Members interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.Member, error)
	Get(ctx context.Context, workspaceSlug, userID string) (entity.Member, error)
}
