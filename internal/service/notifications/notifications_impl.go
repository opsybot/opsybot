package notifications

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx         repository.Transactor
	lock       repository.Lock
	runs       repository.NotificationRun
	rules      repository.NotificationRule
	channels   repository.Channel
	identities repository.ChatIdentity
	alerts     repository.Alert
	workspaces repository.Workspace
	actions    repository.ActionToken
	notifier   service.Notifier
	limiter    service.RateLimiter
	cfg        config.Auth
	chat       config.Chat
}

func New(
	tx repository.Transactor,
	lock repository.Lock,
	runs repository.NotificationRun,
	rules repository.NotificationRule,
	channels repository.Channel,
	identities repository.ChatIdentity,
	alerts repository.Alert,
	workspaces repository.Workspace,
	actions repository.ActionToken,
	notifier service.Notifier,
	limiter service.RateLimiter,
	cfg config.Auth,
	chat config.Chat,
) service.Notifications {
	return &srv{tx: tx, lock: lock, runs: runs, rules: rules, channels: channels, identities: identities, alerts: alerts, workspaces: workspaces, actions: actions, notifier: notifier, limiter: limiter, cfg: cfg, chat: chat}
}

func (s *srv) Page(ctx context.Context, req entity.NotifyRequest) (entity.NotificationRun, error) {
	rule, err := s.rules.Get(ctx, req.WorkspaceID, req.UserID)
	if err != nil {
		if !errors.Is(err, entity.ErrNotificationRuleNotFound) {
			return entity.NotificationRun{}, err
		}
		rule = entity.DefaultNotificationRule(req.WorkspaceID, req.UserID)
	}
	channels, err := s.channels.ListByUser(ctx, req.UserID)
	if err != nil {
		return entity.NotificationRun{}, err
	}
	providers, err := s.identities.LinkedProviders(ctx, req.WorkspaceID, req.UserID)
	if err != nil {
		return entity.NotificationRun{}, err
	}
	plan := entity.BuildNotificationPlan(rule, channels, entity.ChatLinkedChannels(providers), entity.NotifyUrgencyFor(req.Severity), req.Email)
	run := entity.StartNotificationRun(req, plan)
	created, _, err := s.runs.Create(ctx, run)
	if err != nil {
		return entity.NotificationRun{}, err
	}
	return created, nil
}

func (s *srv) StopForAlerts(ctx context.Context, workspaceID string, alertIDs []string, reason entity.NotifyStopReason, now time.Time) error {
	_, err := s.runs.StopByAlerts(ctx, workspaceID, alertIDs, reason, now)
	return err
}

func (s *srv) AttemptsForAlert(ctx context.Context, alertID string) ([]entity.NotificationAttempt, error) {
	return s.runs.ListAttempts(ctx, alertID, entity.AlertTimelineLimit)
}

func (s *srv) Advance(ctx context.Context, now time.Time) (int, error) {
	due, err := s.runs.ListDue(ctx, now, entity.NotificationRunSweepBatch)
	if err != nil {
		return 0, err
	}
	return s.process(ctx, ids(due), now)
}

func (s *srv) RunNow(ctx context.Context, runIDs []string, now time.Time) error {
	_, err := s.process(ctx, runIDs, now)
	return err
}

func (s *srv) process(ctx context.Context, runIDs []string, now time.Time) (int, error) {
	pending := make([]*pendingSend, 0, len(runIDs))
	for _, id := range runIDs {
		p, err := s.executeStep(ctx, id, now)
		if err != nil {
			logger.From(ctx).ErrorContext(ctx, "notification step failed", "run", id, "error", err)
			continue
		}
		if p != nil {
			pending = append(pending, p)
		}
	}
	if len(pending) == 0 {
		return 0, nil
	}
	group, gctx := errgroup.WithContext(ctx)
	group.SetLimit(entity.NotificationSendConcurrency)
	for _, p := range pending {
		p := p
		group.Go(func() error {
			s.deliver(gctx, p, now)
			return nil
		})
	}
	_ = group.Wait()
	return len(pending), nil
}

type pendingSend struct {
	run    entity.NotificationRun
	tick   entity.NotifyStepTick
	target entity.NotifyTarget
	page   entity.AlertPage
}

func (s *srv) executeStep(ctx context.Context, runID string, now time.Time) (*pendingSend, error) {
	var pending *pendingSend
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		held, err := s.lock.TryJob(ctx, "notify:"+runID)
		if err != nil || !held {
			return err
		}
		run, err := s.runs.GetByID(ctx, runID)
		if err != nil {
			return err
		}
		alert, err := s.alerts.GetByID(ctx, run.WorkspaceID, run.AlertID)
		if err != nil {
			return err
		}
		if alert.Status == entity.AlertStatusResolved {
			_, err := s.runs.SaveProgress(ctx, run.Stopped(now, entity.NotifyStopResolved))
			return err
		}
		tick, due := run.NextStep(now)
		if !due {
			return nil
		}
		switch tick.Kind {
		case entity.NotifyStepExhaust:
			finished := run.Finished(now)
			if _, err := s.runs.SaveProgress(ctx, finished); err != nil {
				return err
			}
			if !deliveredAny(ctx, s.runs, run) {
				return s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
					At: now, Kind: entity.AlertEventNotified,
					Text: "Nothing reached " + run.Label, Result: "none delivered",
				})
			}
			return nil
		case entity.NotifyStepSuppress:
			advanced := run.Advanced(now, entity.NotifyOutcomeSuppressed)
			saved, err := s.runs.AdvanceStep(ctx, run.ID, run.StepIndex, advanced)
			if err != nil || !saved {
				return err
			}
			if err := s.runs.AppendAttempt(ctx, s.attempt(run, tick, entity.NotifyOutcomeSuppressed, entity.NotifyResult{Detail: "quiet hours"})); err != nil {
				return err
			}
			if tick.Index == 0 {
				return s.alerts.AppendEvent(ctx, run.AlertID, entity.AlertEvent{
					At: now, Kind: entity.AlertEventPush,
					Text: "Quiet hours held the low-urgency page for " + run.Label, Result: "quiet hours",
				})
			}
			return nil
		case entity.NotifyStepSend:
			if run.StepAttempts >= entity.NotificationStepMaxAttempts {
				gaveUp := run.Advanced(now, entity.NotifyOutcomeFailed)
				if _, err := s.runs.AdvanceStep(ctx, run.ID, run.StepIndex, gaveUp); err != nil {
					return err
				}
				return s.runs.AppendAttempt(ctx, s.attempt(run, tick, entity.NotifyOutcomeFailed, entity.NotifyResult{Detail: "gave up after repeated delivery attempts"}))
			}
			claimed, err := s.runs.Claim(ctx, run.ID, run.StepIndex, now.Add(entity.NotificationStepLeaseTTL))
			if err != nil || !claimed {
				return err
			}
			slug := ""
			if ws, err := s.workspaces.GetByID(ctx, run.WorkspaceID); err == nil {
				slug = ws.Slug
			}
			target := entity.NotifyTarget{
				WorkspaceID: run.WorkspaceID, UserID: run.UserID, Name: run.Label, Channel: tick.Step.Channel,
				ChannelID: tick.Step.ChannelID, Detail: tick.Step.Detail,
			}
			if actionableChannel(tick.Step.Channel) {
				ack, resolve, err := s.mintActions(ctx, run, tick.Step.ChannelID, now)
				if err != nil {
					return err
				}
				target.AckToken = ack
				target.ResolveToken = resolve
			}
			pending = &pendingSend{
				run:    run,
				tick:   tick,
				target: target,
				page:   entity.BuildAlertPage(alert, slug, run.PolicySlug, s.cfg.BaseURL, run.Level),
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pending, nil
}

func (s *srv) deliver(ctx context.Context, p *pendingSend, now time.Time) {
	current, err := s.runs.GetByID(ctx, p.run.ID)
	if err != nil {
		logger.From(ctx).ErrorContext(ctx, "notification deliver reload failed", "run", p.run.ID, "error", err)
		return
	}
	if current.State == entity.NotifyRunStopped {
		_ = s.runs.AppendAttempt(ctx, s.attempt(p.run, p.tick, entity.NotifyOutcomeSkipped, entity.NotifyResult{Detail: "already " + string(current.StopReason)}))
		s.recordEvent(ctx, p, entity.NotifyOutcomeSkipped, entity.NotifyResult{Detail: "already " + string(current.StopReason)})
		return
	}
	allowed, err := s.limiter.Allow(ctx, entity.RateScopeNotify, p.run.UserID)
	if err == nil && !allowed.Allowed {
		_ = s.runs.AppendAttempt(ctx, s.attempt(p.run, p.tick, entity.NotifyOutcomeThrottled, entity.NotifyResult{Detail: "rate limited"}))
		_, _ = s.runs.Reschedule(ctx, p.run.ID, p.tick.Index, now.Add(allowed.RetryAfter))
		return
	}
	target := p.target
	if target.ChannelID != "" && needsSecret(target.Channel) {
		if secret, serr := s.channels.Secret(ctx, target.ChannelID, p.run.UserID); serr == nil {
			target.Secret = secret
		}
	}
	result := s.notifier.Send(ctx, target, p.page)
	outcome := outcomeFor(result)
	saved, err := s.runs.AdvanceStep(ctx, p.run.ID, p.tick.Index, current.Advanced(now, outcome))
	if err != nil {
		logger.From(ctx).ErrorContext(ctx, "advance notification step failed", "run", p.run.ID, "error", err)
		return
	}
	if !saved {
		return
	}
	if err := s.runs.AppendAttempt(ctx, s.attempt(p.run, p.tick, outcome, result)); err != nil {
		logger.From(ctx).ErrorContext(ctx, "record notification attempt failed", "alert", p.run.AlertID, "error", err)
	}
	s.recordEvent(ctx, p, outcome, result)
}

func (s *srv) recordEvent(ctx context.Context, p *pendingSend, outcome entity.NotifyOutcome, result entity.NotifyResult) {
	text := deliveryLabel(p.tick.Step.Channel, p.run.Label)
	if outcome == entity.NotifyOutcomeFailed {
		text = text + " failed"
	}
	if outcome == entity.NotifyOutcomeSkipped {
		text = text + " skipped"
	}
	if outcome == entity.NotifyOutcomeAccepted {
		text = text + " · delivery unconfirmed"
	}
	if err := s.alerts.AppendEvent(ctx, p.run.AlertID, entity.AlertEvent{
		At:     time.Now().UTC(),
		Kind:   p.tick.Step.Channel.EventKind(),
		Text:   text,
		Result: result.Detail,
	}); err != nil {
		logger.From(ctx).ErrorContext(ctx, "record notification event failed", "alert", p.run.AlertID, "error", err)
	}
}

func (s *srv) attempt(run entity.NotificationRun, tick entity.NotifyStepTick, outcome entity.NotifyOutcome, result entity.NotifyResult) entity.NotificationAttempt {
	return entity.NotificationAttempt{
		RunID: run.ID, WorkspaceID: run.WorkspaceID, AlertID: run.AlertID, UserID: run.UserID,
		StepIndex: tick.Index, Channel: tick.Step.Channel, ChannelID: tick.Step.ChannelID,
		Detail: tick.Step.Detail, Outcome: outcome, ProviderMessageID: result.MessageID, ErrorDetail: errorDetail(outcome, result),
	}
}

func outcomeFor(result entity.NotifyResult) entity.NotifyOutcome {
	if !result.Delivered {
		return entity.NotifyOutcomeFailed
	}
	if result.MessageID == "" {
		return entity.NotifyOutcomeAccepted
	}
	return entity.NotifyOutcomeDelivered
}

func errorDetail(outcome entity.NotifyOutcome, result entity.NotifyResult) string {
	if outcome == entity.NotifyOutcomeDelivered || outcome == entity.NotifyOutcomeAccepted {
		return ""
	}
	return result.Detail
}

func deliveryLabel(channel entity.ChannelType, name string) string {
	switch channel {
	case entity.ChannelTypeEmail:
		return "Emailed " + name
	case entity.ChannelTypeNtfy:
		return "ntfy to " + name
	case entity.ChannelTypeWebhook:
		return "Webhook for " + name
	case entity.ChannelTypeSlack:
		return "Slack DM to " + name
	case entity.ChannelTypeTeams:
		return "Teams message to " + name
	case entity.ChannelTypeDiscord:
		return "Discord DM to " + name
	case entity.ChannelTypeTelegram:
		return "Telegram to " + name
	default:
		return "Paged " + name
	}
}

func needsSecret(channel entity.ChannelType) bool {
	return channel == entity.ChannelTypeWebhook || channel == entity.ChannelTypeNtfy
}

func actionableChannel(channel entity.ChannelType) bool {
	return channel.EventKind() == entity.AlertEventChat || channel == entity.ChannelTypeEmail || channel == entity.ChannelTypeNtfy
}

func (s *srv) mintActions(ctx context.Context, run entity.NotificationRun, channelID string, now time.Time) (string, string, error) {
	ttl := s.chat.ActionTokenTTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	expiresAt := now.Add(ttl)
	tokens := map[entity.ActionKind]string{}
	for _, kind := range []entity.ActionKind{entity.ActionKindAck, entity.ActionKindResolve} {
		raw, err := entity.GenerateToken(entity.ActionTokenLength)
		if err != nil {
			return "", "", err
		}
		if err := s.actions.Mint(ctx, entity.AlertActionRecord{
			WorkspaceID: run.WorkspaceID, AlertID: run.AlertID, UserID: run.UserID, ChannelID: channelID,
			Action: kind, TokenHash: entity.HashToken(raw), ExpiresAt: expiresAt,
		}); err != nil {
			return "", "", err
		}
		tokens[kind] = raw
	}
	return tokens[entity.ActionKindAck], tokens[entity.ActionKindResolve], nil
}

func deliveredAny(ctx context.Context, runs repository.NotificationRun, run entity.NotificationRun) bool {
	attempts, err := runs.ListAttempts(ctx, run.AlertID, entity.AlertTimelineLimit)
	if err != nil {
		return true
	}
	for _, a := range attempts {
		if a.RunID == run.ID && (a.Outcome == entity.NotifyOutcomeDelivered || a.Outcome == entity.NotifyOutcomeAccepted) {
			return true
		}
	}
	return false
}

func ids(runs []entity.NotificationRun) []string {
	out := make([]string, len(runs))
	for i, r := range runs {
		out[i] = r.ID
	}
	return out
}
