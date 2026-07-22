package repository

//go:generate go tool mockgen -source=alert_route.go -destination=./alert_route/alert_route_mock.go -package=alert_route

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type AlertRoute interface {
	List(ctx context.Context, workspaceID string) ([]entity.AlertRoute, error)
	Create(ctx context.Context, workspaceID string, in entity.NewAlertRoute) (entity.AlertRoute, error)
	Update(ctx context.Context, workspaceID, routeID string, in entity.NewAlertRoute) (entity.AlertRoute, error)
	Delete(ctx context.Context, workspaceID, routeID string) error
	Reorder(ctx context.Context, workspaceID string, orderedIDs []string) error
	ListGroupRules(ctx context.Context, workspaceID string) ([]entity.GroupRule, error)
	ReplaceGroupRules(ctx context.Context, workspaceID string, rules []entity.GroupRule) error
	Settings(ctx context.Context, workspaceID string) (entity.AlertSettings, error)
	SetDefaultPolicy(ctx context.Context, workspaceID, policyRef string) error
}
