package service

//go:generate go tool mockgen -source=escalations.go -destination=./escalations/escalations_mock.go -package=escalations

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Escalations interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.EscalationPolicySummary, error)
	Get(ctx context.Context, workspaceSlug, policySlug string) (entity.EscalationPolicyDetail, error)
	Create(ctx context.Context, workspaceSlug string, in entity.EscalationPolicy) (entity.EscalationPolicy, error)
	Update(ctx context.Context, workspaceSlug, policySlug string, in entity.EscalationPolicy) (entity.EscalationPolicy, error)
	Delete(ctx context.Context, workspaceSlug, policySlug string) error
	Directory(ctx context.Context, workspaceSlug string) (entity.EscalationDirectory, error)
	ListWebhooks(ctx context.Context, workspaceSlug string) ([]entity.EscalationWebhook, error)
	CreateWebhook(ctx context.Context, workspaceSlug string, in entity.NewEscalationWebhook, secret string) (entity.EscalationWebhook, error)
	UpdateWebhook(ctx context.Context, workspaceSlug, webhookSlug string, in entity.NewEscalationWebhook) (entity.EscalationWebhook, error)
	DeleteWebhook(ctx context.Context, workspaceSlug, webhookSlug string) error
	Escalate(ctx context.Context, workspaceSlug, alertID string) error
	Start(ctx context.Context, alert entity.Alert, policyID string) error
	Advance(ctx context.Context, now time.Time) (int, error)
	OnAcked(ctx context.Context, workspaceID string, alertIDs []string, now time.Time) error
	OnResolved(ctx context.Context, workspaceID string, alertIDs []string, now time.Time) error
	RunForAlert(ctx context.Context, alertID string) (entity.EscalationRun, error)
}
