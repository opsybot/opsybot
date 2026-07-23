package cron

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/service"
)

type NotificationSweep struct {
	notifications service.Notifications
}

func NewNotificationSweep(notifications service.Notifications) *NotificationSweep {
	return &NotificationSweep{notifications: notifications}
}

func (j *NotificationSweep) Run(ctx context.Context) (int, error) {
	return j.notifications.Advance(ctx, time.Now().UTC())
}
