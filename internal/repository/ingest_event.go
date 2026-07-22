package repository

//go:generate go tool mockgen -source=ingest_event.go -destination=./ingest_event/ingest_event_mock.go -package=ingest_event

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type IngestEvent interface {
	Record(ctx context.Context, event entity.IngestEvent) error
	RecordFailure(ctx context.Context, failure entity.IngestFailure) error
	ListFailures(ctx context.Context, workspaceID string, limit int) ([]entity.IngestFailure, error)
	ListBySource(ctx context.Context, sourceID string, limit int) ([]entity.IngestEvent, error)
}
