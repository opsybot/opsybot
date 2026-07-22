package repository

//go:generate go tool mockgen -source=alert_source.go -destination=./alert_source/alert_source_mock.go -package=alert_source

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type AlertSource interface {
	Create(ctx context.Context, workspaceID string, src entity.AlertSource) (entity.AlertSource, error)
	Update(ctx context.Context, workspaceID, slug string, in entity.AlertSourceUpdate) (entity.AlertSource, error)
	Delete(ctx context.Context, workspaceID, slug string) error
	GetBySlug(ctx context.Context, workspaceID, slug string) (entity.AlertSource, error)
	GetByToken(ctx context.Context, token string) (entity.AlertSource, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]entity.AlertSource, error)
	SetPaused(ctx context.Context, workspaceID, slug string, paused bool) error
	RotateSecret(ctx context.Context, workspaceID, slug, secret string) (entity.AlertSource, error)
	ReplaceMappings(ctx context.Context, sourceID string, mappings []entity.SourceMapping) error
	MarkDelivery(ctx context.Context, sourceID string, at time.Time, failed bool) error
}
