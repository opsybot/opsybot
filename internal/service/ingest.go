package service

//go:generate go tool mockgen -source=ingest.go -destination=./ingest/ingest_mock.go -package=ingest

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Ingest interface {
	Webhook(ctx context.Context, req entity.IngestRequest) ([]entity.IngestResult, error)
	CheckIn(ctx context.Context, req entity.CheckInRequest) (entity.IngestResult, error)
	SweepMonitors(ctx context.Context, now time.Time) (int, error)
	ExpireAlerts(ctx context.Context, now time.Time) (int, error)
	PruneIngestHistory(ctx context.Context, now time.Time) (int, error)
}
