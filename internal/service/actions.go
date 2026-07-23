package service

//go:generate go tool mockgen -source=actions.go -destination=./actions/actions_mock.go -package=actions

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Actions interface {
	Redeem(ctx context.Context, rawToken, ip string) (entity.ActionOutcome, error)
}
