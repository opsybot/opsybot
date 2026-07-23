package notifier

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	mailer      repository.Mailer
	pager       repository.Pager
	ntfy        repository.Ntfy
	chatConns   repository.ChatConnection
	chatIDs     repository.ChatIdentity
	chatCourier repository.ChatCourier
	cfg         config.Auth
}

func New(
	mailer repository.Mailer,
	pager repository.Pager,
	ntfy repository.Ntfy,
	chatConns repository.ChatConnection,
	chatIDs repository.ChatIdentity,
	chatCourier repository.ChatCourier,
	cfg config.Auth,
) service.Notifier {
	return &srv{mailer: mailer, pager: pager, ntfy: ntfy, chatConns: chatConns, chatIDs: chatIDs, chatCourier: chatCourier, cfg: cfg}
}

func (s *srv) Send(ctx context.Context, target entity.NotifyTarget, page entity.AlertPage) entity.NotifyResult {
	switch target.Channel {
	case entity.ChannelTypeEmail:
		if target.Detail == "" {
			return entity.NotifyResult{Detail: "no email address on file"}
		}
		if err := s.mailer.SendPage(ctx, target.Detail, page); err != nil {
			return entity.NotifyResult{Detail: err.Error()}
		}
		return entity.NotifyResult{Delivered: true, Detail: "email sent"}
	case entity.ChannelTypeNtfy:
		return s.sendNtfy(ctx, target, page)
	case entity.ChannelTypeWebhook:
		return s.sendWebhook(ctx, target, page)
	case entity.ChannelTypeSlack, entity.ChannelTypeDiscord, entity.ChannelTypeTelegram:
		return s.sendChat(ctx, target, page)
	default:
		return entity.NotifyResult{Detail: "channel not connected yet"}
	}
}

func (s *srv) sendChat(ctx context.Context, target entity.NotifyTarget, page entity.AlertPage) entity.NotifyResult {
	provider := entity.ChatProvider(target.Channel)
	if target.WorkspaceID == "" {
		return entity.NotifyResult{Detail: "no workspace for chat delivery"}
	}
	conn, err := s.chatConns.Get(ctx, target.WorkspaceID, provider)
	if err != nil {
		return entity.NotifyResult{Detail: string(provider) + " is not connected for this workspace"}
	}
	token, err := s.chatConns.BotToken(ctx, target.WorkspaceID, provider)
	if err != nil || token == "" {
		return entity.NotifyResult{Detail: string(provider) + " bot token is unavailable"}
	}
	ident, err := s.chatIDs.GetForUser(ctx, conn.ID, target.UserID)
	if err != nil {
		return entity.NotifyResult{Detail: "no " + string(provider) + " identity for this user"}
	}
	if !ident.Verified {
		return entity.NotifyResult{Detail: string(provider) + " channel is not verified"}
	}
	result, err := s.chatCourier.Send(ctx, entity.ChatDelivery{
		Provider: provider, BotToken: token, ProviderUserID: ident.ProviderUserID, DMChannelID: ident.DMChannelID,
		Page: page, AckToken: target.AckToken, ResolveToken: target.ResolveToken, BaseURL: s.cfg.BaseURL,
	})
	if err != nil {
		return entity.NotifyResult{Detail: err.Error()}
	}
	if result.DMChannelID != "" && result.DMChannelID != ident.DMChannelID {
		_ = s.chatIDs.SetDMChannel(ctx, ident.ID, result.DMChannelID)
	}
	return result.Result
}

func (s *srv) sendNtfy(ctx context.Context, target entity.NotifyTarget, page entity.AlertPage) entity.NotifyResult {
	server, topic := splitNtfy(target.Detail)
	if topic == "" {
		return entity.NotifyResult{Detail: "ntfy topic is missing"}
	}
	result, err := s.ntfy.Publish(ctx, entity.NtfyMessage{
		Server:   server,
		Topic:    topic,
		Token:    target.Secret,
		Title:    page.Subject(),
		Body:     page.PlainText(),
		Priority: ntfyPriority(page.Severity),
		Click:    page.AlertURL,
	})
	if err != nil {
		return entity.NotifyResult{Detail: err.Error()}
	}
	return result
}

func splitNtfy(detail string) (string, string) {
	u, err := url.Parse(detail)
	if err != nil || u.Host == "" {
		return "", strings.Trim(detail, "/")
	}
	topic := strings.Trim(u.Path, "/")
	server := u.Scheme + "://" + u.Host
	return server, topic
}

func ntfyPriority(sev entity.AlertSeverity) int {
	switch sev {
	case entity.SeverityCritical:
		return 5
	case entity.SeverityHigh:
		return 4
	default:
		return 3
	}
}

type channelWebhookPayload struct {
	Event      string    `json:"event"`
	Severity   string    `json:"severity"`
	Service    string    `json:"service"`
	Title      string    `json:"title"`
	PolicySlug string    `json:"policySlug"`
	Level      int       `json:"level"`
	StartedAt  time.Time `json:"startedAt"`
	SentAt     time.Time `json:"sentAt"`
	URL        string    `json:"url"`
}

func (s *srv) sendWebhook(ctx context.Context, target entity.NotifyTarget, page entity.AlertPage) entity.NotifyResult {
	if target.Detail == "" {
		return entity.NotifyResult{Detail: "webhook url is missing"}
	}
	body, err := json.Marshal(channelWebhookPayload{
		Event:      "alert.notified",
		Severity:   string(page.Severity),
		Service:    page.Service,
		Title:      page.Title,
		PolicySlug: page.PolicySlug,
		Level:      page.Level,
		StartedAt:  page.StartedAt.UTC(),
		SentAt:     time.Now().UTC(),
		URL:        page.AlertURL,
	})
	if err != nil {
		return entity.NotifyResult{Detail: err.Error()}
	}
	result, derr := s.pager.DeliverTo(ctx, target.Detail, target.Secret, body)
	if derr != nil {
		return entity.NotifyResult{Detail: derr.Error()}
	}
	return result
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
