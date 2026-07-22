package service

//go:generate go tool mockgen -source=alerts.go -destination=./alerts/alerts_mock.go -package=alerts

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Alerts interface {
	List(ctx context.Context, workspaceSlug string, filter entity.AlertFilter) ([]entity.Alert, string, error)
	Get(ctx context.Context, workspaceSlug, alertID string) (entity.Alert, error)
	Acknowledge(ctx context.Context, workspaceSlug string, ids []string) (int, error)
	Resolve(ctx context.Context, workspaceSlug string, ids []string) (int, error)
	Failures(ctx context.Context, workspaceSlug string, limit int) ([]entity.IngestFailure, error)
}
