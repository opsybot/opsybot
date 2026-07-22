package cron

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/service"
)

type EscalationSweep struct {
	escalations service.Escalations
}

func NewEscalationSweep(escalations service.Escalations) *EscalationSweep {
	return &EscalationSweep{escalations: escalations}
}

func (j *EscalationSweep) Run(ctx context.Context) (int, error) {
	return j.escalations.Advance(ctx, time.Now().UTC())
}
