package repository

//go:generate go tool mockgen -source=escalation_policy.go -destination=./escalation_policy/escalation_policy_mock.go -package=escalation_policy

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type EscalationPolicy interface {
	List(ctx context.Context, workspaceID string) ([]entity.EscalationPolicy, error)
	GetBySlug(ctx context.Context, workspaceID, slug string) (entity.EscalationPolicy, error)
	GetByID(ctx context.Context, workspaceID, id string) (entity.EscalationPolicy, error)
	Create(ctx context.Context, p entity.EscalationPolicy) (entity.EscalationPolicy, error)
	Update(ctx context.Context, p entity.EscalationPolicy) (entity.EscalationPolicy, error)
	Delete(ctx context.Context, workspaceID, id string) error
	Refs(ctx context.Context, workspaceID, policyID string) (entity.EscalationPolicyRefs, error)
	RoutedCounts(ctx context.Context, workspaceID string) (map[string]int, error)
	ListReferencingUser(ctx context.Context, workspaceID, userID string) ([]entity.EscalationPolicy, error)
	ListReferencingSchedule(ctx context.Context, workspaceID, scheduleID string) ([]entity.EscalationPolicy, error)
	ListReferencingTeam(ctx context.Context, workspaceID, teamID string) ([]entity.EscalationPolicy, error)
	ListReferencingWebhook(ctx context.Context, workspaceID, webhookID string) ([]entity.EscalationPolicy, error)
	ListWebhooks(ctx context.Context, workspaceID string) ([]entity.EscalationWebhook, error)
	GetWebhook(ctx context.Context, workspaceID, slug string) (entity.EscalationWebhook, error)
	CreateWebhook(ctx context.Context, workspaceID string, in entity.NewEscalationWebhook, secret string) (entity.EscalationWebhook, error)
	UpdateWebhook(ctx context.Context, workspaceID, slug string, in entity.NewEscalationWebhook) (entity.EscalationWebhook, error)
	DeleteWebhook(ctx context.Context, workspaceID, slug string) error
}
