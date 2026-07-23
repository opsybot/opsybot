package repository

//go:generate go tool mockgen -source=notification_run.go -destination=./notification_run/notification_run_mock.go -package=notification_run

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type NotificationRun interface {
	Create(ctx context.Context, run entity.NotificationRun) (entity.NotificationRun, bool, error)
	GetByID(ctx context.Context, id string) (entity.NotificationRun, error)
	ListDue(ctx context.Context, now time.Time, limit int) ([]entity.NotificationRun, error)
	SaveProgress(ctx context.Context, run entity.NotificationRun) (bool, error)
	Claim(ctx context.Context, runID string, stepIndex int, leasedUntil time.Time) (bool, error)
	AdvanceStep(ctx context.Context, runID string, fromStepIndex int, run entity.NotificationRun) (bool, error)
	Reschedule(ctx context.Context, runID string, stepIndex int, at time.Time) (bool, error)
	StopByAlerts(ctx context.Context, workspaceID string, alertIDs []string, reason entity.NotifyStopReason, at time.Time) (int, error)
	ListByAlert(ctx context.Context, alertID string) ([]entity.NotificationRun, error)
	AppendAttempt(ctx context.Context, attempt entity.NotificationAttempt) error
	ListAttempts(ctx context.Context, alertID string, limit int) ([]entity.NotificationAttempt, error)
}
