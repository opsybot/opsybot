package service

//go:generate go tool mockgen -source=alert_sources.go -destination=./alert_sources/alert_sources_mock.go -package=alert_sources

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type AlertSources interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.AlertSource, error)
	Get(ctx context.Context, workspaceSlug, sourceSlug string) (entity.AlertSource, error)
	Create(ctx context.Context, workspaceSlug string, in entity.NewAlertSource) (entity.AlertSource, error)
	Update(ctx context.Context, workspaceSlug, sourceSlug string, in entity.AlertSourceUpdate) (entity.AlertSource, error)
	Delete(ctx context.Context, workspaceSlug, sourceSlug string) error
	SetPaused(ctx context.Context, workspaceSlug, sourceSlug string, paused bool) error
	RotateSecret(ctx context.Context, workspaceSlug, sourceSlug string) (entity.AlertSource, error)
	SaveMapping(ctx context.Context, workspaceSlug, sourceSlug string, mappings []entity.SourceMapping) (entity.AlertSource, error)
	Events(ctx context.Context, workspaceSlug, sourceSlug string, limit int) ([]entity.IngestEvent, error)
}
