package cron

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/service"
)

type IngestRetention struct {
	ingest service.Ingest
}

func NewIngestRetention(ingest service.Ingest) *IngestRetention {
	return &IngestRetention{ingest: ingest}
}

func (j *IngestRetention) Run(ctx context.Context) (int, error) {
	return j.ingest.PruneIngestHistory(ctx, time.Now().UTC())
}
