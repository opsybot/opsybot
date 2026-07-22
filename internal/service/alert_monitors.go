package service

//go:generate go tool mockgen -source=alert_monitors.go -destination=./alert_monitors/alert_monitors_mock.go -package=alert_monitors

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type AlertMonitors interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.AlertMonitor, error)
	Get(ctx context.Context, workspaceSlug, monitorID string) (entity.AlertMonitor, error)
	Create(ctx context.Context, workspaceSlug string, in entity.NewAlertMonitor) (entity.AlertMonitor, error)
	Update(ctx context.Context, workspaceSlug, monitorID string, in entity.AlertMonitorUpdate) (entity.AlertMonitor, error)
	Delete(ctx context.Context, workspaceSlug, monitorID string) error
}
