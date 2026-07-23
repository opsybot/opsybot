package repository

//go:generate go tool mockgen -source=ntfy.go -destination=./ntfy/ntfy_mock.go -package=ntfy

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Ntfy interface {
	Publish(ctx context.Context, msg entity.NtfyMessage) (entity.NotifyResult, error)
}
