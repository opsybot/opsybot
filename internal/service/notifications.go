package service

//go:generate go tool mockgen -source=notifications.go -destination=./notifications/notifications_mock.go -package=notifications

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Notifications interface {
	Page(ctx context.Context, req entity.NotifyRequest) (entity.NotificationRun, error)
	RunNow(ctx context.Context, runIDs []string, now time.Time) error
	Advance(ctx context.Context, now time.Time) (int, error)
	StopForAlerts(ctx context.Context, workspaceID string, alertIDs []string, reason entity.NotifyStopReason, now time.Time) error
	AttemptsForAlert(ctx context.Context, alertID string) ([]entity.NotificationAttempt, error)
}
