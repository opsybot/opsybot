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
	FindResolved(ctx context.Context, workspaceID, sourceID, dedupKey string, endedAt time.Time) (entity.Alert, error)
	GetByID(ctx context.Context, workspaceID, id string) (entity.Alert, error)
	List(ctx context.Context, workspaceID string, filter entity.AlertFilter) ([]entity.Alert, string, error)
	Facets(ctx context.Context, workspaceID string, since time.Time) (entity.AlertFacets, error)
	Acknowledge(ctx context.Context, workspaceID string, ids []string, userID, label string, at time.Time) (int, error)
	Resolve(ctx context.Context, workspaceID string, ids []string, at time.Time, mode entity.ResolveMode) (int, error)
	Reopen(ctx context.Context, alertID string, at time.Time) (bool, error)
	ApplyRouting(ctx context.Context, alertID, policyRef, groupKey, silenceID string, suppressedAt time.Time) error
	AppendEvent(ctx context.Context, alertID string, event entity.AlertEvent) error
	ReplaceLinks(ctx context.Context, alertID string, links []entity.AlertLink) error
	ListEvents(ctx context.Context, alertID string, limit int) ([]entity.AlertEvent, error)
	ListLinks(ctx context.Context, alertID string) ([]entity.AlertLink, error)
	UpsertGroupParent(ctx context.Context, in entity.AlertUpsert, groupKey string) (entity.Alert, entity.IngestOutcome, error)
	AttachToParent(ctx context.Context, childID, parentID string) error
	DetachFromParent(ctx context.Context, childID string) error
	RollUpParent(ctx context.Context, parentID string, at time.Time) (entity.Alert, error)
	ListChildren(ctx context.Context, parentIDs []string) (map[string][]entity.AlertChild, error)
	ExpireStale(ctx context.Context, now time.Time, limit int) (int, error)
	CountsBySource(ctx context.Context, sourceIDs []string, since time.Time, buckets int) (map[string][]int, error)
}
