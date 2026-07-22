package cron

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/service"
)

type AlertAutoResolve struct {
	ingest service.Ingest
}

func NewAlertAutoResolve(ingest service.Ingest) *AlertAutoResolve {
	return &AlertAutoResolve{ingest: ingest}
}

func (j *AlertAutoResolve) Run(ctx context.Context, now time.Time) (int, error) {
	return j.ingest.ExpireAlerts(ctx, now)
}
