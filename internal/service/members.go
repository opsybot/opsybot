package service

//go:generate go tool mockgen -source=members.go -destination=./members/members_mock.go -package=members

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Members interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.Member, error)
	Get(ctx context.Context, workspaceSlug, userID string) (entity.Member, error)
	Invite(ctx context.Context, workspaceSlug, email string, role entity.Role) (entity.Invite, string, error)
	ListInvites(ctx context.Context, workspaceSlug string) ([]entity.Invite, error)
	ResendInvite(ctx context.Context, workspaceSlug, userID string) (entity.Invite, string, error)
	RevokeInvite(ctx context.Context, workspaceSlug, userID string) error
	ChangeRole(ctx context.Context, workspaceSlug, userID string, role entity.Role) error
	Deactivate(ctx context.Context, workspaceSlug, userID string, replacements map[string]string) error
	Reactivate(ctx context.Context, workspaceSlug, userID string) error
	References(ctx context.Context, workspaceSlug, userID string) ([]entity.MemberReference, error)
}
