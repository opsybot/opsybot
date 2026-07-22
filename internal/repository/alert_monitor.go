package repository

//go:generate go tool mockgen -source=alert_monitor.go -destination=./alert_monitor/alert_monitor_mock.go -package=alert_monitor

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type AlertMonitor interface {
	List(ctx context.Context, workspaceID string) ([]entity.AlertMonitor, error)
	Get(ctx context.Context, workspaceID, monitorID string) (entity.AlertMonitor, error)
	GetBySourceID(ctx context.Context, sourceID string) (entity.AlertMonitor, error)
	Create(ctx context.Context, workspaceID, sourceID string, in entity.NewAlertMonitor) (entity.AlertMonitor, error)
	Update(ctx context.Context, workspaceID, monitorID string, in entity.AlertMonitorUpdate) (entity.AlertMonitor, error)
	RecordCheckIn(ctx context.Context, monitorID string, at time.Time) error
	ListDue(ctx context.Context, now time.Time, limit int) ([]entity.AlertMonitor, error)
}
