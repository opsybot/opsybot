package notifier

import (
	"context"
	"encoding/json"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	mailer repository.Mailer
	pager  repository.Pager
}

func New(mailer repository.Mailer, pager repository.Pager) service.Notifier {
	return &srv{mailer: mailer, pager: pager}
}

func (s *srv) PageUser(ctx context.Context, member entity.Member, page entity.AlertPage) entity.NotifyResult {
	if member.Email == "" {
		return entity.NotifyResult{Detail: "no email address on file"}
	}
	if err := s.mailer.SendPage(ctx, member.Email, page); err != nil {
		return entity.NotifyResult{Detail: err.Error()}
	}
	return entity.NotifyResult{Delivered: true, Detail: "email sent"}
}

type webhookAlertPayload struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Severity  string            `json:"severity"`
	Status    string            `json:"status"`
	Source    string            `json:"source"`
	Service   string            `json:"service"`
	Labels    map[string]string `json:"labels"`
	StartedAt time.Time         `json:"startedAt"`
	URL       string            `json:"url"`
}

type webhookPayload struct {
	Event      string              `json:"event"`
	PolicySlug string              `json:"policySlug"`
	Level      int                 `json:"level"`
	SentAt     time.Time           `json:"sentAt"`
	Alert      webhookAlertPayload `json:"alert"`
}

func (s *srv) CallWebhook(ctx context.Context, hook entity.EscalationWebhook, alert entity.Alert, page entity.AlertPage) entity.NotifyResult {
	body, err := json.Marshal(webhookPayload{
		Event:      "alert.escalated",
		PolicySlug: page.PolicySlug,
		Level:      page.Level,
		SentAt:     time.Now().UTC(),
		Alert: webhookAlertPayload{
			ID:        alert.ID,
			Title:     alert.Title,
			Severity:  string(alert.Severity),
			Status:    string(alert.Status),
			Source:    alert.SourceLabel,
			Service:   alert.ServiceName,
			Labels:    alert.Labels,
			StartedAt: alert.StartedAt.UTC(),
			URL:       page.AlertURL,
		},
	})
	if err != nil {
		return entity.NotifyResult{Detail: err.Error()}
	}
	result, err := s.pager.Deliver(ctx, hook, body)
	if err != nil {
		return entity.NotifyResult{Detail: err.Error()}
	}
	return result
}
