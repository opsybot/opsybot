package service

//go:generate go tool mockgen -source=notification_rules.go -destination=./notification_rules/notification_rules_mock.go -package=notification_rules

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type NotificationRules interface {
	Get(ctx context.Context, workspaceSlug string) (entity.NotificationSettings, error)
	Save(ctx context.Context, workspaceSlug string, in entity.NotificationRule) (entity.NotificationRule, error)
}
