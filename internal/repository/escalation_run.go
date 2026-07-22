package repository

//go:generate go tool mockgen -source=escalation_run.go -destination=./escalation_run/escalation_run_mock.go -package=escalation_run

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type EscalationRun interface {
	Create(ctx context.Context, run entity.EscalationRun) (entity.EscalationRun, bool, error)
	GetByAlertID(ctx context.Context, alertID string) (entity.EscalationRun, error)
	ListDue(ctx context.Context, now time.Time, limit int) ([]entity.EscalationRun, error)
	SaveProgress(ctx context.Context, run entity.EscalationRun) (bool, error)
	MarkAcked(ctx context.Context, alertID string, ackedAt, expiresAt time.Time) (bool, error)
	MarkResolved(ctx context.Context, alertIDs []string, at time.Time) (int, error)
	Resume(ctx context.Context, runID string, at time.Time) (bool, error)
	Finish(ctx context.Context, runID string, state entity.EscalationRunState, at time.Time) (bool, error)
	NextRoundRobin(ctx context.Context, policyID, nodeID string) (int, error)
	RecentByPolicy(ctx context.Context, policyID string, limit int) ([]entity.EscalationRecent, error)
}
