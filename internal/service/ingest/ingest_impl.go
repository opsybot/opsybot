package ingest

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx       repository.Transactor
	sources  repository.AlertSource
	alerts   repository.Alert
	events   repository.IngestEvent
	routes   repository.AlertRoute
	silences repository.Silence
	monitors repository.AlertMonitor
	limiter  repository.RateLimiter
	lock     repository.Lock
	cfg      config.Ingest
}

func New(
	tx repository.Transactor,
	sources repository.AlertSource,
	alerts repository.Alert,
	events repository.IngestEvent,
	routes repository.AlertRoute,
	silences repository.Silence,
	monitors repository.AlertMonitor,
	limiter repository.RateLimiter,
	lock repository.Lock,
	cfg config.Ingest,
) service.Ingest {
	return &srv{
		tx: tx, sources: sources, alerts: alerts, events: events, routes: routes,
		silences: silences, monitors: monitors, limiter: limiter, lock: lock, cfg: cfg,
	}
}

type routingContext struct {
	routes     []entity.AlertRoute
	groupRules []entity.GroupRule
	silences   []entity.Silence
	defaultRef string
}

func (s *srv) routingContext(ctx context.Context, workspaceID string, now time.Time) (routingContext, error) {
	routes, err := s.routes.List(ctx, workspaceID)
	if err != nil {
		return routingContext{}, err
	}
	groupRules, err := s.routes.ListGroupRules(ctx, workspaceID)
	if err != nil {
		return routingContext{}, err
	}
	silences, err := s.silences.ListActive(ctx, workspaceID, now)
	if err != nil {
		return routingContext{}, err
	}
	settings, err := s.routes.Settings(ctx, workspaceID)
	if err != nil {
		return routingContext{}, err
	}
	return routingContext{routes: routes, groupRules: groupRules, silences: silences, defaultRef: settings.DefaultPolicyRef}, nil
}

func (s *srv) route(ctx context.Context, rc routingContext, alert entity.Alert, upsert entity.AlertUpsert, now time.Time, policyOverride string) error {
	rule, groupKey, grouped := entity.GroupKeyFor(rc.groupRules, alert)
	if !grouped || strings.HasPrefix(alert.DedupKey, entity.GroupDedupPrefix) {
		return s.applyRouting(ctx, rc, alert, now, policyOverride)
	}

	parent, outcome, err := s.alerts.UpsertGroupParent(ctx, entity.GroupParentFor(rule, groupKey, upsert, alert), groupKey)
	if err != nil {
		return err
	}
	if err := s.alerts.AttachToParent(ctx, alert.ID, parent.ID); err != nil {
		return err
	}
	if err := s.alerts.ApplyRouting(ctx, alert.ID, "", groupKey, "", now); err != nil {
		return err
	}
	if err := s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
		At:     now,
		Kind:   entity.AlertEventGrouped,
		Text:   "Grouped into " + parent.Title,
		Result: groupKey,
	}); err != nil {
		return err
	}
	if outcome == entity.IngestOutcomeCreated {
		if err := s.alerts.AppendEvent(ctx, parent.ID, entity.AlertEvent{
			At:   now,
			Kind: entity.AlertEventGrouped,
			Text: rule.Describes(),
		}); err != nil {
			return err
		}
	}
	rolled, err := s.alerts.RollUpParent(ctx, parent.ID, now)
	if err != nil {
		return err
	}
	return s.applyRouting(ctx, rc, rolled, now, policyOverride)
}

func (s *srv) applyRouting(ctx context.Context, rc routingContext, alert entity.Alert, now time.Time, policyOverride string) error {
	fallback := rc.defaultRef
	if policyOverride != "" {
		fallback = policyOverride
	}
	_, policyRef, _ := entity.RouteFor(rc.routes, alert, fallback)
	silence, suppressed := entity.SilenceFor(rc.silences, alert, now)

	silenceID := ""
	if suppressed {
		silenceID = silence.ID
	}
	if err := s.alerts.ApplyRouting(ctx, alert.ID, policyRef, alert.GroupKey, silenceID, now); err != nil {
		return err
	}
	if err := s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
		At:     now,
		Kind:   entity.AlertEventRouted,
		Text:   "Routed to " + policyRef,
		Result: alert.GroupKey,
	}); err != nil {
		return err
	}
	if suppressed {
		return s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
			At:   now,
			Kind: entity.AlertEventSuppressed,
			Text: "Suppressed by an active silence. Still recorded, but it pages no one.",
		})
	}
	return nil
}

func (s *srv) Webhook(ctx context.Context, req entity.IngestRequest) ([]entity.IngestResult, error) {
	src, err := s.sources.GetByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	if src.Paused {
		return nil, entity.ErrAlertSourcePaused
	}

	now := req.ReceivedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	if len(req.Body) == 0 {
		detail := "The request body was empty."
		return nil, s.reject(ctx, src, entity.FailureEmptyBody, detail, "",
			entity.ParseFailure(entity.FailureEmptyBody, detail))
	}
	if int64(len(req.Body)) > s.cfg.MaxBodyBytes {
		detail := fmt.Sprintf("The body exceeded %d bytes.", s.cfg.MaxBodyBytes)
		return nil, s.reject(ctx, src, entity.FailureBodyTooLarge, detail, "",
			entity.ParseFailure(entity.FailureBodyTooLarge, detail))
	}
	if err := s.verifySignature(ctx, src, req, now); err != nil {
		return nil, err
	}
	if err := s.guardFlood(ctx, src, now); err != nil {
		return nil, err
	}

	parsed, err := parseFor(src.Format, req.Body, src, now)
	if err != nil {
		if pe, ok := entity.ParseFailureOf(err); ok {
			return nil, s.reject(ctx, src, pe.Reason, pe.Detail, string(req.Body), err)
		}
		return nil, err
	}

	results := make([]entity.IngestResult, 0, len(parsed))
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		rc, err := s.routingContext(ctx, src.WorkspaceID, now)
		if err != nil {
			return err
		}
		for _, raw := range parsed {
			normalized := raw.Normalize(src, now)
			if !normalized.Valid() {
				continue
			}
			result, err := s.apply(ctx, src, normalized, now, rc, "")
			if err != nil {
				return err
			}
			results = append(results, result)
		}
		return s.sources.MarkDelivery(ctx, src.ID, now, false)
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *srv) apply(ctx context.Context, src entity.AlertSource, in entity.IngestedAlert, now time.Time, rc routingContext, policyOverride string) (entity.IngestResult, error) {
	dedupKey := entity.DeriveDedupKey(src.ID, in.DedupKeyRaw, in.Title, in.SourceLabel, in.Labels)

	upsert := entity.AlertUpsert{
		WorkspaceID: src.WorkspaceID,
		SourceID:    src.ID,
		DedupKey:    dedupKey,
		Title:       in.Title,
		Description: in.Description,
		Severity:    in.Severity,
		SourceLabel: in.SourceLabel,
		ServiceName: in.ServiceName,
		Labels:      in.Labels,
		StartedAt:   in.StartedAt,
		LastSeenAt:  latest(in.StartedAt, in.EndedAt, now),
		Payload:     in.Payload,
		Links:       in.Links,
	}

	if in.Resolved {
		return s.applyResolve(ctx, src, upsert, in, dedupKey)
	}

	alert, outcome, err := s.alerts.UpsertOpen(ctx, upsert)
	if err != nil {
		return entity.IngestResult{}, err
	}
	if outcome == entity.IngestOutcomeUpdated && in.Repeat {
		outcome = entity.IngestOutcomeDuplicate
	}

	if outcome == entity.IngestOutcomeCreated {
		if err := s.alerts.ReplaceLinks(ctx, alert.ID, in.Links); err != nil {
			return entity.IngestResult{}, err
		}
		if err := s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
			At:   now,
			Kind: entity.AlertEventReceived,
			Text: fmt.Sprintf("Alert received from %s", src.Slug),
		}); err != nil {
			return entity.IngestResult{}, err
		}
	} else if err := s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
		At:     now,
		Kind:   entity.AlertEventDeduped,
		Text:   fmt.Sprintf("Repeat received from %s", src.Slug),
		Result: fmt.Sprintf("count %d", alert.Count),
	}); err != nil {
		return entity.IngestResult{}, err
	}

	if err := s.route(ctx, rc, alert, upsert, now, policyOverride); err != nil {
		return entity.IngestResult{}, err
	}
	return s.finish(ctx, src, alert.ID, dedupKey, outcome, now)
}

func (s *srv) applyResolve(ctx context.Context, src entity.AlertSource, upsert entity.AlertUpsert, in entity.IngestedAlert, dedupKey string) (entity.IngestResult, error) {
	endedAt := in.EndedAt
	if endedAt.IsZero() {
		endedAt = upsert.LastSeenAt
	}

	alert, outcome, err := s.alerts.ResolveByDedupKey(ctx, src.WorkspaceID, src.ID, dedupKey, endedAt, in.ResolveMode)
	switch {
	case errors.Is(err, entity.ErrAlertNotFound):
		if existing, findErr := s.alerts.FindResolved(ctx, src.WorkspaceID, src.ID, dedupKey, endedAt); findErr == nil {
			return s.finish(ctx, src, existing.ID, dedupKey, entity.IngestOutcomeDuplicate, endedAt)
		} else if !errors.Is(findErr, entity.ErrAlertNotFound) {
			return entity.IngestResult{}, findErr
		}
		created, insertErr := s.alerts.InsertResolved(ctx, upsert, endedAt, in.ResolveMode)
		if insertErr != nil {
			return entity.IngestResult{}, insertErr
		}
		if err := s.alerts.AppendEvent(ctx, created.ID, entity.AlertEvent{
			At:   endedAt,
			Kind: entity.AlertEventResolved,
			Text: "Resolved on arrival: no matching open alert",
		}); err != nil {
			return entity.IngestResult{}, err
		}
		return s.finish(ctx, src, created.ID, dedupKey, entity.IngestOutcomeResolved, endedAt)
	case err != nil:
		return entity.IngestResult{}, err
	}

	if outcome == entity.IngestOutcomeStale {
		return s.finish(ctx, src, alert.ID, dedupKey, outcome, endedAt)
	}
	if err := s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
		At:   endedAt,
		Kind: entity.AlertEventResolved,
		Text: "Source reported recovery",
	}); err != nil {
		return entity.IngestResult{}, err
	}
	if alert.ParentAlertID != "" {
		if _, err := s.alerts.RollUpParent(ctx, alert.ParentAlertID, endedAt); err != nil {
			return entity.IngestResult{}, err
		}
	}
	return s.finish(ctx, src, alert.ID, dedupKey, outcome, endedAt)
}

func (s *srv) finish(ctx context.Context, src entity.AlertSource, alertID, dedupKey string, outcome entity.IngestOutcome, at time.Time) (entity.IngestResult, error) {
	if err := s.events.Record(ctx, entity.IngestEvent{
		WorkspaceID: src.WorkspaceID,
		SourceID:    src.ID,
		AlertID:     alertID,
		DedupKey:    dedupKey,
		Outcome:     outcome,
		At:          at,
	}); err != nil {
		return entity.IngestResult{}, err
	}
	return entity.IngestResult{AlertID: alertID, DedupKey: dedupKey, Outcome: outcome}, nil
}

func (s *srv) guardFlood(ctx context.Context, src entity.AlertSource, now time.Time) error {
	if s.cfg.RatePerMin <= 0 {
		return nil
	}
	limit := entity.RateLimit{Rate: s.cfg.RatePerMin, Period: time.Minute, Burst: s.cfg.RatePerMin}
	result, err := s.limiter.Allow(ctx, string(entity.RateScopeIngest)+":"+src.ID, limit)
	if err != nil {
		return err
	}
	if result.Allowed {
		return nil
	}

	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		alert, outcome, err := s.alerts.UpsertOpen(ctx, entity.FloodAlert(src, s.cfg.RatePerMin, now))
		if err != nil {
			return err
		}
		if outcome == entity.IngestOutcomeCreated {
			if err := s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
				At:   now,
				Kind: entity.AlertEventReceived,
				Text: "Ingest budget exceeded, so further events from this source are being dropped.",
			}); err != nil {
				return err
			}
		}
		return s.events.Record(ctx, entity.IngestEvent{
			WorkspaceID: src.WorkspaceID,
			SourceID:    src.ID,
			AlertID:     alert.ID,
			DedupKey:    alert.DedupKey,
			Outcome:     entity.IngestOutcomeFloodDropped,
			At:          now,
		})
	})
	if err != nil {
		return err
	}
	return entity.ErrIngestFlooded
}

func (s *srv) CheckIn(ctx context.Context, req entity.CheckInRequest) (entity.IngestResult, error) {
	src, err := s.sources.GetByToken(ctx, req.Token)
	if err != nil {
		return entity.IngestResult{}, err
	}
	if src.Format != entity.SourceFormatHeartbeat {
		return entity.IngestResult{}, entity.ErrAlertMonitorFormat
	}
	if src.Paused {
		return entity.IngestResult{}, entity.ErrAlertSourcePaused
	}

	now := req.ReceivedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	monitor, err := s.monitors.GetBySourceID(ctx, src.ID)
	if err != nil {
		return entity.IngestResult{}, err
	}

	result := entity.IngestResult{DedupKey: monitor.DedupKey(), Outcome: entity.IngestOutcomeDuplicate}
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.monitors.RecordCheckIn(ctx, monitor.ID, now); err != nil {
			return err
		}
		alert, outcome, err := s.alerts.ResolveByDedupKey(ctx, src.WorkspaceID, src.ID, monitor.DedupKey(), now, entity.ResolveModeSource)
		switch {
		case errors.Is(err, entity.ErrAlertNotFound):
		case err != nil:
			return err
		case outcome == entity.IngestOutcomeResolved:
			if err := s.alerts.AppendEvent(ctx, alert.ID, entity.AlertEvent{
				At:   now,
				Kind: entity.AlertEventResolved,
				Text: monitor.Name + " checked in again",
			}); err != nil {
				return err
			}
			result = entity.IngestResult{AlertID: alert.ID, DedupKey: monitor.DedupKey(), Outcome: outcome}
		}
		if err := s.sources.MarkDelivery(ctx, src.ID, now, false); err != nil {
			return err
		}
		return s.events.Record(ctx, entity.IngestEvent{
			WorkspaceID: src.WorkspaceID,
			SourceID:    src.ID,
			AlertID:     result.AlertID,
			DedupKey:    result.DedupKey,
			Outcome:     result.Outcome,
			At:          now,
		})
	})
	if err != nil {
		return entity.IngestResult{}, err
	}
	return result, nil
}

func (s *srv) SweepMonitors(ctx context.Context, now time.Time) (int, error) {
	due, err := s.monitors.ListDue(ctx, now, entity.MonitorSweepBatch)
	if err != nil {
		return 0, err
	}

	fired := 0
	for _, monitor := range due {
		raised, err := s.fireMonitor(ctx, monitor, now)
		if err != nil {
			return fired, err
		}
		if raised {
			fired++
		}
	}
	return fired, nil
}

func (s *srv) fireMonitor(ctx context.Context, monitor entity.AlertMonitor, now time.Time) (bool, error) {
	raised := false
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		held, err := s.lock.TryJob(ctx, "monitor:"+monitor.ID)
		if err != nil || !held {
			return err
		}
		src, err := s.sources.GetBySlug(ctx, monitor.WorkspaceID, monitor.Slug)
		if err != nil {
			return err
		}
		rc, err := s.routingContext(ctx, monitor.WorkspaceID, now)
		if err != nil {
			return err
		}
		if _, err := s.apply(ctx, src, entity.MonitorAlert(monitor, src, now).Normalize(src, now), now, rc, monitor.PolicyRef); err != nil {
			return err
		}
		raised = true
		return nil
	})
	return raised, err
}

func (s *srv) ExpireAlerts(ctx context.Context, now time.Time) (int, error) {
	return s.alerts.ExpireStale(ctx, now, entity.AlertExpireBatch)
}

func (s *srv) PruneIngestHistory(ctx context.Context, now time.Time) (int, error) {
	if s.cfg.FailureRetention <= 0 {
		return 0, nil
	}
	return s.events.Prune(ctx, now.Add(-s.cfg.FailureRetention))
}

func (s *srv) verifySignature(ctx context.Context, src entity.AlertSource, req entity.IngestRequest, now time.Time) error {
	secrets := src.SecretsInGrace(now)
	if req.Signature == "" && !src.RequireSignature {
		return nil
	}
	if req.Signature == "" || len(secrets) == 0 || !entity.VerifyBodySignature(secrets, req.Body, req.Signature) {
		return s.reject(ctx, src, entity.FailureSignatureInvalid,
			"The request signature did not match the source secret.", string(req.Body),
			entity.ErrAlertSourceSignature)
	}
	return nil
}

func (s *srv) recordFailure(ctx context.Context, src entity.AlertSource, reason entity.IngestFailureReason, detail, payload string) error {
	failure := entity.IngestFailure{
		WorkspaceID: src.WorkspaceID,
		SourceID:    src.ID,
		Reason:      reason,
		Detail:      detail,
		Payload:     payload,
		At:          time.Now().UTC(),
	}
	if err := s.events.RecordFailure(ctx, failure); err != nil {
		return err
	}
	return s.sources.MarkDelivery(ctx, src.ID, failure.At, true)
}

func (s *srv) reject(ctx context.Context, src entity.AlertSource, reason entity.IngestFailureReason, detail, payload string, domainErr error) error {
	if err := s.recordFailure(ctx, src, reason, detail, payload); err != nil {
		return err
	}
	return domainErr
}

func latest(values ...time.Time) time.Time {
	var out time.Time
	for _, v := range values {
		if v.After(out) {
			out = v
		}
	}
	return out
}
