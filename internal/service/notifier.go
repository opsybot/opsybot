package service

//go:generate go tool mockgen -source=notifier.go -destination=./notifier/notifier_mock.go -package=notifier

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Notifier interface {
	Send(ctx context.Context, target entity.NotifyTarget, page entity.AlertPage) entity.NotifyResult
	CallWebhook(ctx context.Context, hook entity.EscalationWebhook, alert entity.Alert, page entity.AlertPage) entity.NotifyResult
}
