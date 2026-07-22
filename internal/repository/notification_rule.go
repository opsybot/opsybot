package repository

//go:generate go tool mockgen -source=notification_rule.go -destination=./notification_rule/notification_rule_mock.go -package=notification_rule

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type NotificationRule interface {
	Get(ctx context.Context, workspaceID, userID string) (entity.NotificationRule, error)
	Save(ctx context.Context, rule entity.NotificationRule) (entity.NotificationRule, error)
	DeleteByUser(ctx context.Context, workspaceID, userID string) error
}
