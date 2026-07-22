package service

//go:generate go tool mockgen -source=alert_routes.go -destination=./alert_routes/alert_routes_mock.go -package=alert_routes

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type AlertRoutes interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.AlertRoute, entity.AlertSettings, error)
	Create(ctx context.Context, workspaceSlug string, in entity.NewAlertRoute) (entity.AlertRoute, error)
	Update(ctx context.Context, workspaceSlug, routeID string, in entity.NewAlertRoute) (entity.AlertRoute, error)
	Delete(ctx context.Context, workspaceSlug, routeID string) error
	Reorder(ctx context.Context, workspaceSlug string, ids []string) error
	SetDefaultPolicy(ctx context.Context, workspaceSlug, policyRef string) error
	Preview(ctx context.Context, workspaceSlug, payload string) (entity.RoutePreview, error)
	ListGroupRules(ctx context.Context, workspaceSlug string) ([]entity.GroupRule, error)
	SaveGroupRules(ctx context.Context, workspaceSlug string, rules []entity.GroupRule) ([]entity.GroupRule, error)
}
