package ingest_event

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.IngestEvent {
	return &repo{db: db}
}

func (r *repo) Record(ctx context.Context, event entity.IngestEvent) error {
	at := event.At
	if at.IsZero() {
		at = time.Now().UTC()
	}
	m := &dbpostgres.AlertIngestEvent{
		WorkspaceID: event.WorkspaceID,
		SourceID:    event.SourceID,
		DedupKey:    event.DedupKey,
		Outcome:     string(event.Outcome),
		At:          at,
	}
	if event.AlertID != "" {
		m.AlertID = null.StringFrom(event.AlertID)
	}
	cols := boil.Whitelist("workspace_id", "source_id", "alert_id", "dedup_key", "outcome", "at")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		return fmt.Errorf("record ingest event: %w", err)
	}
	return nil
}

func (r *repo) RecordFailure(ctx context.Context, failure entity.IngestFailure) error {
	at := failure.At
	if at.IsZero() {
		at = time.Now().UTC()
	}
	m := &dbpostgres.AlertIngestFailure{
		WorkspaceID: failure.WorkspaceID,
		Reason:      string(failure.Reason),
		Detail:      failure.Detail,
		Payload:     failure.Payload,
		At:          at,
	}
	if failure.SourceID != "" {
		m.SourceID = null.StringFrom(failure.SourceID)
	}
	cols := boil.Whitelist("workspace_id", "source_id", "reason", "detail", "payload", "at")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		return fmt.Errorf("record ingest failure: %w", err)
	}
	return nil
}

func (r *repo) ListFailures(ctx context.Context, workspaceID string, limit int) ([]entity.IngestFailure, error) {
	if limit <= 0 || limit > entity.AlertListMaxPageSize {
		limit = entity.AlertListDefaultPageSize
	}
	rows, err := dbpostgres.AlertIngestFailures(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("at DESC, id DESC"),
		qm.Limit(limit),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list ingest failures: %w", err)
	}
	out := make([]entity.IngestFailure, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.IngestFailure{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			SourceID:    m.SourceID.String,
			Reason:      entity.IngestFailureReason(m.Reason),
			Detail:      m.Detail,
			Payload:     m.Payload,
			At:          m.At,
		})
	}
	return out, nil
}

func (r *repo) Prune(ctx context.Context, before time.Time) (int, error) {
	exec := r.db.Querier(ctx)
	failures, err := dbpostgres.AlertIngestFailures(qm.Where("at < ?", before)).DeleteAll(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("prune ingest failures: %w", err)
	}
	events, err := dbpostgres.AlertIngestEvents(qm.Where("at < ?", before)).DeleteAll(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("prune ingest events: %w", err)
	}
	return int(failures + events), nil
}

func (r *repo) ListBySource(ctx context.Context, sourceID string, limit int) ([]entity.IngestEvent, error) {
	if limit <= 0 || limit > entity.AlertListMaxPageSize {
		limit = entity.AlertListDefaultPageSize
	}
	rows, err := dbpostgres.AlertIngestEvents(
		qm.Where("source_id = ?", sourceID),
		qm.OrderBy("at DESC, id DESC"),
		qm.Limit(limit),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list ingest events: %w", err)
	}
	out := make([]entity.IngestEvent, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.IngestEvent{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			SourceID:    m.SourceID,
			AlertID:     m.AlertID.String,
			DedupKey:    m.DedupKey,
			Outcome:     entity.IngestOutcome(m.Outcome),
			At:          m.At,
		})
	}
	return out, nil
}
