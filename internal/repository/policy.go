package repository

//go:generate go tool mockgen -source=policy.go -destination=./policy/policy_mock.go -package=policy

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Policy interface {
	Allowed(ctx context.Context, subject, workspaceID string, obj entity.PolicyObject, act entity.PolicyAction) (bool, error)
	RoleOf(ctx context.Context, userID, workspaceID string) (entity.Role, bool, error)
	RolesByWorkspace(ctx context.Context, workspaceID string) (map[string]entity.Role, error)
	AssignRole(ctx context.Context, userID, workspaceID string, role entity.Role) error
	ReplaceRole(ctx context.Context, userID, workspaceID string, from, to entity.Role) error
	RemoveRole(ctx context.Context, userID, workspaceID string) error
	SeedWorkspace(ctx context.Context, workspaceID string) error
	AssignAgentRole(ctx context.Context, workspaceID string, role entity.Role) error
	RoleOfTx(ctx context.Context, userID, workspaceID string) (entity.Role, bool, error)
	CountActiveAdminsTx(ctx context.Context, workspaceID string) (int, error)
}
