package alert_monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const selectMonitorSQL = `
SELECT m.id, m.workspace_id, m.source_id, m.interval_seconds, m.grace_seconds, m.policy_ref, m.severity,
       m.last_check_in_at, m.created_at, m.updated_at,
       s.slug, s.name, s.ingest_token, s.paused_at
FROM alert_monitors m
JOIN alert_sources s ON s.id = m.source_id`

const listMonitorsSQL = selectMonitorSQL + `
WHERE m.workspace_id = $1
ORDER BY s.name, m.id`

const getMonitorSQL = selectMonitorSQL + `
WHERE m.workspace_id = $1 AND m.id = $2`

const getMonitorBySourceSQL = selectMonitorSQL + `
WHERE m.source_id = $1`

const listDueMonitorsSQL = selectMonitorSQL + `
WHERE s.paused_at IS NULL
  AND COALESCE(m.last_check_in_at, m.created_at) + make_interval(secs => m.interval_seconds + m.grace_seconds) < $1
  AND NOT EXISTS (
      SELECT 1 FROM alerts a
      WHERE a.source_id = m.source_id AND a.dedup_key = $2 || m.id::text AND a.resolved_at IS NULL
  )
ORDER BY m.id
LIMIT $3`

type row struct {
	ID              string    `boil:"id"`
	WorkspaceID     string    `boil:"workspace_id"`
	SourceID        string    `boil:"source_id"`
	IntervalSeconds int       `boil:"interval_seconds"`
	GraceSeconds    int       `boil:"grace_seconds"`
	PolicyRef       string    `boil:"policy_ref"`
	Severity        string    `boil:"severity"`
	LastCheckInAt   null.Time `boil:"last_check_in_at"`
	CreatedAt       time.Time `boil:"created_at"`
	UpdatedAt       time.Time `boil:"updated_at"`
	Slug            string    `boil:"slug"`
	Name            string    `boil:"name"`
	IngestToken     string    `boil:"ingest_token"`
	PausedAt        null.Time `boil:"paused_at"`
}

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.AlertMonitor {
	return &repo{db: db}
}

func toEntity(r row) entity.AlertMonitor {
	return entity.AlertMonitor{
		ID:            r.ID,
		WorkspaceID:   r.WorkspaceID,
		SourceID:      r.SourceID,
		Slug:          r.Slug,
		Name:          r.Name,
		Interval:      time.Duration(r.IntervalSeconds) * time.Second,
		Grace:         time.Duration(r.GraceSeconds) * time.Second,
		PolicyRef:     r.PolicyRef,
		Severity:      entity.AlertSeverity(r.Severity),
		LastCheckInAt: r.LastCheckInAt.Time,
		Paused:        r.PausedAt.Valid,
		CheckInToken:  r.IngestToken,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func (r *repo) query(ctx context.Context, sql string, args ...any) ([]entity.AlertMonitor, error) {
	var rows []row
	if err := queries.Raw(sql, args...).Bind(ctx, r.db.Querier(ctx), &rows); err != nil {
		return nil, err
	}
	out := make([]entity.AlertMonitor, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEntity(m))
	}
	return out, nil
}

func (r *repo) one(ctx context.Context, sql string, args ...any) (entity.AlertMonitor, error) {
	found, err := r.query(ctx, sql, args...)
	if err != nil {
		return entity.AlertMonitor{}, fmt.Errorf("get alert monitor: %w", err)
	}
	if len(found) == 0 {
		return entity.AlertMonitor{}, entity.ErrAlertMonitorNotFound
	}
	return found[0], nil
}

func (r *repo) List(ctx context.Context, workspaceID string) ([]entity.AlertMonitor, error) {
	out, err := r.query(ctx, listMonitorsSQL, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list alert monitors: %w", err)
	}
	return out, nil
}

func (r *repo) Get(ctx context.Context, workspaceID, monitorID string) (entity.AlertMonitor, error) {
	return r.one(ctx, getMonitorSQL, workspaceID, monitorID)
}

func (r *repo) GetBySourceID(ctx context.Context, sourceID string) (entity.AlertMonitor, error) {
	return r.one(ctx, getMonitorBySourceSQL, sourceID)
}

func (r *repo) Create(ctx context.Context, workspaceID, sourceID string, in entity.NewAlertMonitor) (entity.AlertMonitor, error) {
	m := &dbpostgres.AlertMonitor{
		WorkspaceID:     workspaceID,
		SourceID:        sourceID,
		IntervalSeconds: int(in.Interval / time.Second),
		GraceSeconds:    int(in.Grace / time.Second),
		PolicyRef:       in.PolicyRef,
		Severity:        string(in.Severity),
	}
	cols := boil.Whitelist("workspace_id", "source_id", "interval_seconds", "grace_seconds", "policy_ref", "severity")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		return entity.AlertMonitor{}, fmt.Errorf("create alert monitor: %w", err)
	}
	return r.Get(ctx, workspaceID, m.ID)
}

func (r *repo) Update(ctx context.Context, workspaceID, monitorID string, in entity.AlertMonitorUpdate) (entity.AlertMonitor, error) {
	values := dbpostgres.M{
		"interval_seconds": int(in.Interval / time.Second),
		"grace_seconds":    int(in.Grace / time.Second),
		"policy_ref":       in.PolicyRef,
		"severity":         string(in.Severity),
		"updated_at":       time.Now().UTC(),
	}
	affected, err := dbpostgres.AlertMonitors(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, monitorID),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return entity.AlertMonitor{}, fmt.Errorf("update alert monitor: %w", err)
	}
	if affected == 0 {
		return entity.AlertMonitor{}, entity.ErrAlertMonitorNotFound
	}
	return r.Get(ctx, workspaceID, monitorID)
}

func (r *repo) RecordCheckIn(ctx context.Context, monitorID string, at time.Time) error {
	values := dbpostgres.M{"last_check_in_at": at, "updated_at": at}
	if _, err := dbpostgres.AlertMonitors(qm.Where("id = ?", monitorID)).UpdateAll(ctx, r.db.Querier(ctx), values); err != nil {
		return fmt.Errorf("record monitor check-in: %w", err)
	}
	return nil
}

func (r *repo) ListDue(ctx context.Context, now time.Time, limit int) ([]entity.AlertMonitor, error) {
	if limit <= 0 {
		limit = entity.MonitorSweepBatch
	}
	out, err := r.query(ctx, listDueMonitorsSQL, now, entity.MonitorDedupPrefix, limit)
	if err != nil {
		return nil, fmt.Errorf("list due monitors: %w", err)
	}
	return out, nil
}
