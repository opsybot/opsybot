package repository

//go:generate go tool mockgen -source=alert.go -destination=./alert/alert_mock.go -package=alert

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Alert interface {
	UpsertOpen(ctx context.Context, in entity.AlertUpsert) (entity.Alert, entity.IngestOutcome, error)
	ResolveByDedupKey(ctx context.Context, workspaceID, sourceID, dedupKey string, endedAt time.Time, mode entity.ResolveMode) (entity.Alert, entity.IngestOutcome, error)
	InsertResolved(ctx context.Context, in entity.AlertUpsert, endedAt time.Time, mode entity.ResolveMode) (entity.Alert, error)
	GetByID(ctx context.Context, workspaceID, id string) (entity.Alert, error)
	List(ctx context.Context, workspaceID string, filter entity.AlertFilter) ([]entity.Alert, string, error)
	Acknowledge(ctx context.Context, workspaceID string, ids []string, userID, label string, at time.Time) (int, error)
	Resolve(ctx context.Context, workspaceID string, ids []string, at time.Time, mode entity.ResolveMode) (int, error)
	AppendEvent(ctx context.Context, alertID string, event entity.AlertEvent) error
	ReplaceLinks(ctx context.Context, alertID string, links []entity.AlertLink) error
	ListEvents(ctx context.Context, alertID string, limit int) ([]entity.AlertEvent, error)
	ListLinks(ctx context.Context, alertID string) ([]entity.AlertLink, error)
}
