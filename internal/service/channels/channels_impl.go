package channels

import (
	"context"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	channels      repository.Channel
	verifications repository.ChannelVerification
	mailer        repository.Mailer
	ntfy          repository.Ntfy
	pager         repository.Pager
	notifier      service.Notifier
	limiter       service.RateLimiter
	audit         repository.Audit
	cfg           config.Auth
}

func New(
	channels repository.Channel,
	verifications repository.ChannelVerification,
	mailer repository.Mailer,
	ntfy repository.Ntfy,
	pager repository.Pager,
	notifier service.Notifier,
	limiter service.RateLimiter,
	audit repository.Audit,
	cfg config.Auth,
) service.Channels {
	return &srv{
		channels: channels, verifications: verifications, mailer: mailer, ntfy: ntfy,
		pager: pager, notifier: notifier, limiter: limiter, audit: audit, cfg: cfg,
	}
}

func (s *srv) userID(ctx context.Context) (string, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok || id.Kind != entity.IdentityKindSession {
		return "", entity.ErrUnauthenticated
	}
	return id.UserID, nil
}

func (s *srv) List(ctx context.Context) ([]entity.Channel, error) {
	userID, err := s.userID(ctx)
	if err != nil {
		return nil, err
	}
	return s.channels.ListByUser(ctx, userID)
}

func (s *srv) Add(ctx context.Context, in entity.NewChannel) (entity.Channel, error) {
	userID, err := s.userID(ctx)
	if err != nil {
		return entity.Channel{}, err
	}
	if err := in.Validate(); err != nil {
		return entity.Channel{}, err
	}
	ch, err := s.channels.Create(ctx, userID, in)
	if err != nil {
		return entity.Channel{}, err
	}
	_ = s.audit.Create(ctx, entity.AuditEvent{
		ActorType: entity.AuditActorUser, ActorUserID: userID,
		Action: entity.ActionChannelAdded, Target: string(ch.Type),
	})
	return ch, nil
}

func (s *srv) StartVerification(ctx context.Context, channelID string) (entity.ChannelVerification, error) {
	userID, err := s.userID(ctx)
	if err != nil {
		return entity.ChannelVerification{}, err
	}
	ch, err := s.channels.Get(ctx, channelID, userID)
	if err != nil {
		return entity.ChannelVerification{}, err
	}
	token, err := entity.GenerateToken(32)
	if err != nil {
		return entity.ChannelVerification{}, err
	}
	code, err := entity.GenerateNumericCode()
	if err != nil {
		return entity.ChannelVerification{}, err
	}
	nonce, err := entity.GenerateToken(16)
	if err != nil {
		return entity.ChannelVerification{}, err
	}
	method := ch.Type.VerifyMethod()
	ttl := entity.ChannelVerifyTTL
	if ch.Type == entity.ChannelTypeEmail {
		ttl = entity.ChannelVerifyEmailTTL
	}
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	if err := s.verifications.Start(ctx, entity.ChannelVerifyRecord{
		ChannelID: ch.ID, UserID: userID, Method: method,
		TokenHash: entity.HashToken(token), CodeHash: entity.HashToken(code), Nonce: nonce, ExpiresAt: expiresAt,
	}); err != nil {
		return entity.ChannelVerification{}, err
	}
	verifyURL := strings.TrimRight(s.cfg.BaseURL, "/") + "/v1/channels/verify/" + token
	out := entity.ChannelVerification{Method: method, ExpiresAt: expiresAt}
	switch ch.Type {
	case entity.ChannelTypeEmail:
		body := "Confirm this address for Opsybot alerts.\n\nYour code is " + code + "\nOr open: " + verifyURL + "\n\nThis expires in 24 hours."
		if err := s.mailer.SendText(ctx, ch.Detail, "Verify your Opsybot channel", body); err != nil {
			return entity.ChannelVerification{}, err
		}
		out.Detail = "Check your email for a code or confirmation link."
	case entity.ChannelTypeNtfy:
		server, topic := splitNtfy(ch.Detail)
		secret, _ := s.channels.Secret(ctx, ch.ID, userID)
		if _, err := s.ntfy.Publish(ctx, entity.NtfyMessage{
			Server: server, Topic: topic, Token: secret,
			Title: "Verify your Opsybot channel", Body: "Your code is " + code + ". Confirm: " + verifyURL,
			Priority: 3, Click: verifyURL,
		}); err != nil {
			return entity.ChannelVerification{}, err
		}
		out.Detail = "Check the ntfy push for a code or tap Confirm."
	case entity.ChannelTypeWebhook:
		secret, _ := s.channels.Secret(ctx, ch.ID, userID)
		payload := []byte(`{"type":"opsybot.verification","nonce":"` + nonce + `","echo":"` + verifyURL + `"}`)
		result, _ := s.pager.DeliverTo(ctx, ch.Detail, secret, payload)
		if !result.Delivered {
			out.Detail = "Sent a challenge. Your endpoint must POST back to the echo URL to confirm."
		} else {
			out.Detail = "Challenge delivered. Confirm by POSTing back to the echo URL."
		}
	default:
		out.Detail = "Verification for this channel is not available yet."
	}
	return out, nil
}

func (s *srv) CompleteVerification(ctx context.Context, channelID, code string) error {
	userID, err := s.userID(ctx)
	if err != nil {
		return err
	}
	claim, err := s.verifications.ConsumeCode(ctx, channelID, userID, entity.HashToken(strings.TrimSpace(code)), time.Now().UTC())
	if err != nil {
		return err
	}
	return s.markVerified(ctx, claim)
}

func (s *srv) CompleteByToken(ctx context.Context, token string) error {
	claim, err := s.verifications.ConsumeToken(ctx, entity.HashToken(strings.TrimSpace(token)), time.Now().UTC())
	if err != nil {
		return err
	}
	return s.markVerified(ctx, claim)
}

func (s *srv) markVerified(ctx context.Context, claim entity.ChannelVerifyClaim) error {
	if err := s.channels.MarkVerified(ctx, claim.ChannelID, claim.UserID); err != nil {
		return err
	}
	_ = s.audit.Create(ctx, entity.AuditEvent{
		ActorType: entity.AuditActorUser, ActorUserID: claim.UserID,
		Action: entity.ActionChannelVerified, Target: claim.ChannelID,
	})
	return nil
}

func (s *srv) SendTest(ctx context.Context, channelID string) (entity.NotifyResult, error) {
	userID, err := s.userID(ctx)
	if err != nil {
		return entity.NotifyResult{}, err
	}
	ch, err := s.channels.Get(ctx, channelID, userID)
	if err != nil {
		return entity.NotifyResult{}, err
	}
	if !ch.Verified {
		return entity.NotifyResult{}, entity.ErrChannelNotVerified
	}
	allowed, err := s.limiter.Allow(ctx, entity.RateScopeChannelTest, userID)
	if err == nil && !allowed.Allowed {
		return entity.NotifyResult{}, entity.ErrRateLimited
	}
	secret, _ := s.channels.Secret(ctx, ch.ID, userID)
	target := entity.NotifyTarget{
		UserID: userID, Name: "you", Channel: ch.Type, ChannelID: ch.ID, Detail: ch.Detail, Secret: secret,
	}
	page := entity.AlertPage{
		Severity: entity.SeverityHigh, Service: "opsybot", Title: "Test notification",
		StartedAt: time.Now().UTC(), PolicySlug: "test", Level: 1,
		AlertURL: strings.TrimRight(s.cfg.BaseURL, "/"),
	}
	result := s.notifier.Send(ctx, target, page)
	_ = s.audit.Create(ctx, entity.AuditEvent{
		ActorType: entity.AuditActorUser, ActorUserID: userID,
		Action: entity.ActionChannelTested, Target: string(ch.Type),
	})
	return result, nil
}

func (s *srv) Remove(ctx context.Context, channelID string) error {
	userID, err := s.userID(ctx)
	if err != nil {
		return err
	}
	if err := s.channels.Delete(ctx, channelID, userID); err != nil {
		return err
	}
	_ = s.audit.Create(ctx, entity.AuditEvent{
		ActorType: entity.AuditActorUser, ActorUserID: userID,
		Action: entity.ActionChannelRemoved, Target: channelID,
	})
	return nil
}

func splitNtfy(detail string) (string, string) {
	trimmed := strings.TrimRight(detail, "/")
	idx := strings.LastIndex(trimmed, "/")
	if idx < 0 {
		return "", trimmed
	}
	return trimmed[:idx], trimmed[idx+1:]
}
