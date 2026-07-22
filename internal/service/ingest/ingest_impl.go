package ingest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx      repository.Transactor
	sources repository.AlertSource
	alerts  repository.Alert
	events  repository.IngestEvent
	cfg     config.Ingest
}

func New(
	tx repository.Transactor,
	sources repository.AlertSource,
	alerts repository.Alert,
	events repository.IngestEvent,
	cfg config.Ingest,
) service.Ingest {
	return &srv{tx: tx, sources: sources, alerts: alerts, events: events, cfg: cfg}
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

	parsed, err := parseFor(src.Format, req.Body, src, now)
	if err != nil {
		if pe, ok := entity.ParseFailureOf(err); ok {
			return nil, s.reject(ctx, src, pe.Reason, pe.Detail, string(req.Body), err)
		}
		return nil, err
	}

	results := make([]entity.IngestResult, 0, len(parsed))
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		for _, raw := range parsed {
			normalized := raw.Normalize(src, now)
			if !normalized.Valid() {
				continue
			}
			result, err := s.apply(ctx, src, normalized, now)
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

func (s *srv) apply(ctx context.Context, src entity.AlertSource, in entity.IngestedAlert, now time.Time) (entity.IngestResult, error) {
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
