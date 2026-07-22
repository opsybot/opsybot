package cron

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/service"
)

type HeartbeatSweep struct {
	ingest service.Ingest
}

func NewHeartbeatSweep(ingest service.Ingest) *HeartbeatSweep {
	return &HeartbeatSweep{ingest: ingest}
}

func (j *HeartbeatSweep) Run(ctx context.Context) (int, error) {
	return j.ingest.SweepMonitors(ctx, time.Now().UTC())
}
