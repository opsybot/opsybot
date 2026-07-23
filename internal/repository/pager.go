package repository

//go:generate go tool mockgen -source=pager.go -destination=./pager/pager_mock.go -package=pager

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Pager interface {
	Deliver(ctx context.Context, hook entity.EscalationWebhook, payload []byte) (entity.NotifyResult, error)
	DeliverTo(ctx context.Context, url, secret string, payload []byte) (entity.NotifyResult, error)
}
