package alert

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const upsertOpenSQL = `
INSERT INTO alerts (workspace_id, source_id, dedup_key, title, description, severity, status,
                    source_label, service_name, labels, count, started_at, last_seen_at, payload)
VALUES ($1, $2, $3, $4, $5, $6, 'open', $7, $8, $9, 1, $10, $11, $12)
ON CONFLICT (workspace_id, source_id, dedup_key) WHERE resolved_at IS NULL
DO UPDATE SET
    count        = alerts.count + 1,
    last_seen_at = GREATEST(alerts.last_seen_at, EXCLUDED.last_seen_at),
    started_at   = LEAST(alerts.started_at, EXCLUDED.started_at),
    title        = CASE WHEN EXCLUDED.last_seen_at >= alerts.last_seen_at THEN EXCLUDED.title ELSE alerts.title END,
    description  = CASE WHEN EXCLUDED.last_seen_at >= alerts.last_seen_at THEN EXCLUDED.description ELSE alerts.description END,
    severity     = CASE WHEN EXCLUDED.last_seen_at >= alerts.last_seen_at THEN EXCLUDED.severity ELSE alerts.severity END,
    labels       = CASE WHEN EXCLUDED.last_seen_at >= alerts.last_seen_at THEN EXCLUDED.labels ELSE alerts.labels END,
    service_name = CASE WHEN EXCLUDED.last_seen_at >= alerts.last_seen_at THEN EXCLUDED.service_name ELSE alerts.service_name END,
    payload      = CASE WHEN EXCLUDED.last_seen_at >= alerts.last_seen_at THEN EXCLUDED.payload ELSE alerts.payload END,
    updated_at   = now()
RETURNING id, (xmax = 0) AS inserted`

const upsertGroupParentSQL = `
INSERT INTO alerts (workspace_id, source_id, dedup_key, group_key, title, description, severity, status,
                    source_label, service_name, labels, count, started_at, last_seen_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'open', $8, $9, $10, 1, $11, $12)
ON CONFLICT (workspace_id, group_key) WHERE group_key <> '' AND parent_alert_id IS NULL AND resolved_at IS NULL
DO UPDATE SET
    last_seen_at = GREATEST(alerts.last_seen_at, EXCLUDED.last_seen_at),
    started_at   = LEAST(alerts.started_at, EXCLUDED.started_at),
    updated_at   = now()
RETURNING id, (xmax = 0) AS inserted`

const rollUpParentSQL = `
WITH agg AS (
    SELECT count(*) AS all_children,
           count(*) FILTER (WHERE resolved_at IS NULL) AS open_children,
           COALESCE(sum(count) FILTER (WHERE resolved_at IS NULL), 0) AS open_count,
           max(last_seen_at) AS last_seen,
           min(started_at) AS started,
           min(CASE severity WHEN 'critical' THEN 0 WHEN 'high' THEN 1 ELSE 2 END)
               FILTER (WHERE resolved_at IS NULL) AS severity_rank
    FROM alerts
    WHERE parent_alert_id = $1
)
UPDATE alerts p SET
    count        = GREATEST(1, agg.open_count),
    started_at   = LEAST(p.started_at, COALESCE(agg.started, p.started_at)),
    last_seen_at = GREATEST(p.last_seen_at, COALESCE(agg.last_seen, p.last_seen_at)),
    severity     = CASE agg.severity_rank WHEN 0 THEN 'critical' WHEN 1 THEN 'high' WHEN 2 THEN 'warning' ELSE p.severity END,
    status       = CASE WHEN agg.all_children > 0 AND agg.open_children = 0 THEN 'resolved' ELSE p.status END,
    resolved_at  = CASE WHEN agg.all_children > 0 AND agg.open_children = 0 THEN COALESCE(p.resolved_at, $2) ELSE p.resolved_at END,
    ended_at     = CASE WHEN agg.all_children > 0 AND agg.open_children = 0 THEN COALESCE(p.ended_at, $2) ELSE p.ended_at END,
    resolve_mode = CASE WHEN agg.all_children > 0 AND agg.open_children = 0 THEN 'source' ELSE p.resolve_mode END,
    updated_at   = $2
FROM agg
WHERE p.id = $1`

const expireStaleSQL = `
WITH due AS (
    SELECT a.id
    FROM alerts a
    JOIN alert_sources s ON s.id = a.source_id
    WHERE a.resolved_at IS NULL
      AND s.auto_resolve_after_seconds > 0
      AND a.last_seen_at + make_interval(secs => s.auto_resolve_after_seconds) < $1
    ORDER BY a.last_seen_at
    LIMIT $2
)
UPDATE alerts SET
    status       = 'resolved',
    resolved_at  = $1,
    ended_at     = $1,
    resolve_mode = 'timeout',
    updated_at   = $1
WHERE id IN (SELECT id FROM due)`

const facetsSQL = `
WITH scoped AS (
    SELECT source_label, service_name, labels
    FROM alerts
    WHERE workspace_id = $1 AND parent_alert_id IS NULL AND last_seen_at >= $2
)
(SELECT 'source' AS kind, source_label AS value FROM scoped WHERE source_label <> '' GROUP BY 2 ORDER BY 2 LIMIT $3)
UNION ALL
(SELECT 'service', service_name FROM scoped WHERE service_name <> '' GROUP BY 2 ORDER BY 2 LIMIT $3)
UNION ALL
(SELECT 'label', label.key || ':' || label.value
 FROM scoped, LATERAL jsonb_each_text(scoped.labels) AS label
 GROUP BY 2 ORDER BY 2 LIMIT $3)`

const countsBySourceSQL = `
SELECT source_id,
       LEAST($4::int, GREATEST(0, floor(extract(epoch FROM (at - $2::timestamptz)) / $3::float)::int)) AS bucket,
       count(*)::int AS total
FROM alert_ingest_events
WHERE source_id = ANY($1) AND at >= $2::timestamptz
GROUP BY 1, 2`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Alert {
	return &repo{db: db}
}

func marshalLabels(labels map[string]string) (types.JSON, error) {
	if labels == nil {
		labels = map[string]string{}
	}
	raw, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("encode alert labels: %w", err)
	}
	return types.JSON(raw), nil
}

func unmarshalLabels(raw types.JSON) map[string]string {
	out := map[string]string{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func toEntity(m *dbpostgres.Alert) entity.Alert {
	return entity.Alert{
		ID:                    m.ID,
		WorkspaceID:           m.WorkspaceID,
		SourceID:              m.SourceID,
		ParentAlertID:         m.ParentAlertID.String,
		DedupKey:              m.DedupKey,
		GroupKey:              m.GroupKey,
		Title:                 m.Title,
		Description:           m.Description,
		Severity:              entity.AlertSeverity(m.Severity),
		Status:                entity.AlertStatus(m.Status),
		SourceLabel:           m.SourceLabel,
		ServiceName:           m.ServiceName,
		Labels:                unmarshalLabels(m.Labels),
		Count:                 m.Count,
		StartedAt:             m.StartedAt,
		LastSeenAt:            m.LastSeenAt,
		EndedAt:               m.EndedAt.Time,
		AckedAt:               m.AckedAt.Time,
		ResolvedAt:            m.ResolvedAt.Time,
		AckedByUserID:         m.AckedByUserID.String,
		AckedByLabel:          m.AckedByLabel,
		ResolveMode:           entity.ResolveMode(m.ResolveMode),
		RoutedPolicyRef:       m.RoutedPolicyRef,
		SuppressedBySilenceID: m.SuppressedBySilenceID.String,
		SuppressedAt:          m.SuppressedAt.Time,
		Payload:               m.Payload,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}

func (r *repo) UpsertOpen(ctx context.Context, in entity.AlertUpsert) (entity.Alert, entity.IngestOutcome, error) {
	labels, err := marshalLabels(in.Labels)
	if err != nil {
		return entity.Alert{}, entity.IngestOutcomeFailed, err
	}

	var row struct {
		ID       string `boil:"id"`
		Inserted bool   `boil:"inserted"`
	}
	err = queries.Raw(upsertOpenSQL,
		in.WorkspaceID, in.SourceID, in.DedupKey, in.Title, in.Description, string(in.Severity),
		in.SourceLabel, in.ServiceName, labels, in.StartedAt, in.LastSeenAt, in.Payload,
	).Bind(ctx, r.db.Querier(ctx), &row)
	if err != nil {
		return entity.Alert{}, entity.IngestOutcomeFailed, fmt.Errorf("upsert alert: %w", err)
	}

	m, err := dbpostgres.FindAlert(ctx, r.db.Querier(ctx), row.ID)
	if err != nil {
		return entity.Alert{}, entity.IngestOutcomeFailed, fmt.Errorf("reload alert: %w", err)
	}

	outcome := entity.IngestOutcomeUpdated
	if row.Inserted {
		outcome = entity.IngestOutcomeCreated
	}
	return toEntity(m), outcome, nil
}

func (r *repo) ResolveByDedupKey(ctx context.Context, workspaceID, sourceID, dedupKey string, endedAt time.Time, mode entity.ResolveMode) (entity.Alert, entity.IngestOutcome, error) {
	m, err := dbpostgres.Alerts(
		qm.Where("workspace_id = ? AND source_id = ? AND dedup_key = ? AND resolved_at IS NULL", workspaceID, sourceID, dedupKey),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Alert{}, entity.IngestOutcomeStale, entity.ErrAlertNotFound
		}
		return entity.Alert{}, entity.IngestOutcomeFailed, fmt.Errorf("find open alert: %w", err)
	}
	if endedAt.Before(m.StartedAt) {
		return toEntity(m), entity.IngestOutcomeStale, nil
	}

	m.Status = string(entity.AlertStatusResolved)
	m.ResolvedAt = null.TimeFrom(endedAt)
	m.EndedAt = null.TimeFrom(endedAt)
	m.ResolveMode = string(mode)
	m.LastSeenAt = maxTime(m.LastSeenAt, endedAt)
	m.UpdatedAt = time.Now().UTC()
	cols := boil.Whitelist("status", "resolved_at", "ended_at", "resolve_mode", "last_seen_at", "updated_at")
	if _, err := m.Update(ctx, r.db.Querier(ctx), cols); err != nil {
		return entity.Alert{}, entity.IngestOutcomeFailed, fmt.Errorf("resolve alert: %w", err)
	}
	return toEntity(m), entity.IngestOutcomeResolved, nil
}

func (r *repo) InsertResolved(ctx context.Context, in entity.AlertUpsert, endedAt time.Time, mode entity.ResolveMode) (entity.Alert, error) {
	labels, err := marshalLabels(in.Labels)
	if err != nil {
		return entity.Alert{}, err
	}
	m := &dbpostgres.Alert{
		WorkspaceID: in.WorkspaceID,
		SourceID:    in.SourceID,
		DedupKey:    in.DedupKey,
		Title:       in.Title,
		Description: in.Description,
		Severity:    string(in.Severity),
		Status:      string(entity.AlertStatusResolved),
		SourceLabel: in.SourceLabel,
		ServiceName: in.ServiceName,
		Labels:      labels,
		Count:       1,
		StartedAt:   in.StartedAt,
		LastSeenAt:  maxTime(in.LastSeenAt, endedAt),
		EndedAt:     null.TimeFrom(endedAt),
		ResolvedAt:  null.TimeFrom(endedAt),
		ResolveMode: string(mode),
		Payload:     in.Payload,
	}
	cols := boil.Whitelist("workspace_id", "source_id", "dedup_key", "title", "description", "severity", "status",
		"source_label", "service_name", "labels", "count", "started_at", "last_seen_at", "ended_at", "resolved_at",
		"resolve_mode", "payload")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		return entity.Alert{}, fmt.Errorf("insert resolved alert: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) FindResolved(ctx context.Context, workspaceID, sourceID, dedupKey string, endedAt time.Time) (entity.Alert, error) {
	m, err := dbpostgres.Alerts(
		qm.Where("workspace_id = ? AND source_id = ? AND dedup_key = ? AND resolved_at IS NOT NULL AND ended_at = ?",
			workspaceID, sourceID, dedupKey, endedAt),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Alert{}, entity.ErrAlertNotFound
		}
		return entity.Alert{}, fmt.Errorf("find resolved alert: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetByID(ctx context.Context, workspaceID, id string) (entity.Alert, error) {
	m, err := dbpostgres.Alerts(qm.Where("workspace_id = ? AND id = ?", workspaceID, id)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Alert{}, entity.ErrAlertNotFound
		}
		return entity.Alert{}, fmt.Errorf("get alert: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) List(ctx context.Context, workspaceID string, filter entity.AlertFilter) ([]entity.Alert, string, error) {
	limit := filter.Limit
	if limit <= 0 || limit > entity.AlertListMaxPageSize {
		limit = entity.AlertListDefaultPageSize
	}

	mods := []qm.QueryMod{qm.Where("workspace_id = ? AND parent_alert_id IS NULL", workspaceID)}
	if len(filter.Statuses) > 0 {
		mods = append(mods, qm.WhereIn("status IN ?", anySlice(statusStrings(filter.Statuses))...))
	}
	if len(filter.Severities) > 0 {
		mods = append(mods, qm.WhereIn("severity IN ?", anySlice(severityStrings(filter.Severities))...))
	}
	if len(filter.SourceIDs) > 0 {
		mods = append(mods, qm.WhereIn("source_id IN ?", anySlice(filter.SourceIDs)...))
	}
	if len(filter.Sources) > 0 {
		mods = append(mods, qm.WhereIn("source_label IN ?", anySlice(filter.Sources)...))
	}
	if len(filter.Services) > 0 {
		mods = append(mods, qm.WhereIn("service_name IN ?", anySlice(filter.Services)...))
	}
	for _, label := range filter.Labels {
		key, value, ok := strings.Cut(label, ":")
		if !ok {
			continue
		}
		mods = append(mods, qm.Where("labels ->> ? = ?", key, value))
	}
	if !filter.Since.IsZero() {
		mods = append(mods, qm.Where("last_seen_at >= ?", filter.Since))
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		mods = append(mods, qm.Where("(lower(title) LIKE ? OR lower(service_name) LIKE ? OR lower(source_label) LIKE ?)", like, like, like))
	}
	if cursorAt, cursorID, ok := decodeCursor(filter.Cursor); ok {
		mods = append(mods, qm.Where("(last_seen_at, id) < (?, ?)", cursorAt, cursorID))
	}
	mods = append(mods, qm.OrderBy("last_seen_at DESC, id DESC"), qm.Limit(limit+1))

	rows, err := dbpostgres.Alerts(mods...).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, "", fmt.Errorf("list alerts: %w", err)
	}

	next := ""
	if len(rows) > limit {
		last := rows[limit-1]
		next = encodeCursor(last.LastSeenAt, last.ID)
		rows = rows[:limit]
	}

	out := make([]entity.Alert, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEntity(m))
	}
	return out, next, nil
}

func (r *repo) Acknowledge(ctx context.Context, workspaceID string, ids []string, userID, label string, at time.Time) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	values := dbpostgres.M{
		"status":         string(entity.AlertStatusAcked),
		"acked_at":       at,
		"acked_by_label": label,
		"updated_at":     at,
	}
	if userID != "" {
		values["acked_by_user_id"] = userID
	}
	affected, err := dbpostgres.Alerts(
		qm.Where("workspace_id = ? AND status = ?", workspaceID, string(entity.AlertStatusOpen)),
		qm.WhereIn("id IN ?", anySlice(ids)...),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return 0, fmt.Errorf("acknowledge alerts: %w", err)
	}
	return int(affected), nil
}

func (r *repo) Resolve(ctx context.Context, workspaceID string, ids []string, at time.Time, mode entity.ResolveMode) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	values := dbpostgres.M{
		"status":       string(entity.AlertStatusResolved),
		"resolved_at":  at,
		"ended_at":     at,
		"resolve_mode": string(mode),
		"updated_at":   at,
	}
	affected, err := dbpostgres.Alerts(
		qm.Where("workspace_id = ? AND resolved_at IS NULL", workspaceID),
		qm.WhereIn("id IN ?", anySlice(ids)...),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return 0, fmt.Errorf("resolve alerts: %w", err)
	}
	return int(affected), nil
}

func (r *repo) AppendEvent(ctx context.Context, alertID string, event entity.AlertEvent) error {
	at := event.At
	if at.IsZero() {
		at = time.Now().UTC()
	}
	m := &dbpostgres.AlertEvent{
		AlertID: alertID,
		At:      at,
		Kind:    string(event.Kind),
		Text:    event.Text,
		Result:  event.Result,
	}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("alert_id", "at", "kind", "text", "result")); err != nil {
		return fmt.Errorf("append alert event: %w", err)
	}
	return nil
}

func (r *repo) ReplaceLinks(ctx context.Context, alertID string, links []entity.AlertLink) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.AlertLinks(qm.Where("alert_id = ?", alertID)).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("clear alert links: %w", err)
	}
	for i, link := range links {
		m := &dbpostgres.AlertLink{
			AlertID:  alertID,
			Kind:     string(link.Kind),
			Label:    link.Label,
			URL:      link.URL,
			Position: i,
		}
		if err := m.Insert(ctx, exec, boil.Whitelist("alert_id", "kind", "label", "url", "position")); err != nil {
			return fmt.Errorf("insert alert link: %w", err)
		}
	}
	return nil
}

func (r *repo) ListEvents(ctx context.Context, alertID string, limit int) ([]entity.AlertEvent, error) {
	if limit <= 0 || limit > entity.AlertTimelineLimit {
		limit = entity.AlertTimelineLimit
	}
	rows, err := dbpostgres.AlertEvents(
		qm.Where("alert_id = ?", alertID),
		qm.OrderBy("at, id"),
		qm.Limit(limit),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list alert events: %w", err)
	}
	out := make([]entity.AlertEvent, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.AlertEvent{
			ID:      m.ID,
			AlertID: m.AlertID,
			At:      m.At,
			Kind:    entity.AlertEventKind(m.Kind),
			Text:    m.Text,
			Result:  m.Result,
		})
	}
	return out, nil
}

func (r *repo) ListLinks(ctx context.Context, alertID string) ([]entity.AlertLink, error) {
	rows, err := dbpostgres.AlertLinks(
		qm.Where("alert_id = ?", alertID),
		qm.OrderBy("position, id"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list alert links: %w", err)
	}
	out := make([]entity.AlertLink, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.AlertLink{Kind: entity.AlertLinkKind(m.Kind), Label: m.Label, URL: m.URL})
	}
	return out, nil
}

func maxTime(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}

func anySlice(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}

func statusStrings(in []entity.AlertStatus) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = string(s)
	}
	return out
}

func severityStrings(in []entity.AlertSeverity) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = string(s)
	}
	return out
}

func encodeCursor(at time.Time, id string) string {
	return at.UTC().Format(time.RFC3339Nano) + "|" + id
}

func decodeCursor(cursor string) (time.Time, string, bool) {
	at, id, ok := strings.Cut(strings.TrimSpace(cursor), "|")
	if !ok || id == "" {
		return time.Time{}, "", false
	}
	parsed, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return time.Time{}, "", false
	}
	return parsed, id, true
}

func (r *repo) Facets(ctx context.Context, workspaceID string, since time.Time) (entity.AlertFacets, error) {
	var rows []struct {
		Kind  string `boil:"kind"`
		Value string `boil:"value"`
	}
	err := queries.Raw(facetsSQL, workspaceID, since, entity.AlertFacetLimit).Bind(ctx, r.db.Querier(ctx), &rows)
	if err != nil {
		return entity.AlertFacets{}, fmt.Errorf("list alert facets: %w", err)
	}

	out := entity.AlertFacets{Sources: []string{}, Services: []string{}, Labels: []string{}}
	for _, row := range rows {
		switch row.Kind {
		case "source":
			out.Sources = append(out.Sources, row.Value)
		case "service":
			out.Services = append(out.Services, row.Value)
		case "label":
			out.Labels = append(out.Labels, row.Value)
		}
	}
	return out, nil
}

func (r *repo) UpsertGroupParent(ctx context.Context, in entity.AlertUpsert, groupKey string) (entity.Alert, entity.IngestOutcome, error) {
	labels, err := marshalLabels(in.Labels)
	if err != nil {
		return entity.Alert{}, entity.IngestOutcomeFailed, err
	}

	var row struct {
		ID       string `boil:"id"`
		Inserted bool   `boil:"inserted"`
	}
	err = queries.Raw(upsertGroupParentSQL,
		in.WorkspaceID, in.SourceID, in.DedupKey, groupKey, in.Title, in.Description, string(in.Severity),
		in.SourceLabel, in.ServiceName, labels, in.StartedAt, in.LastSeenAt,
	).Bind(ctx, r.db.Querier(ctx), &row)
	if err != nil {
		return entity.Alert{}, entity.IngestOutcomeFailed, fmt.Errorf("upsert group parent: %w", err)
	}

	m, err := dbpostgres.FindAlert(ctx, r.db.Querier(ctx), row.ID)
	if err != nil {
		return entity.Alert{}, entity.IngestOutcomeFailed, fmt.Errorf("reload group parent: %w", err)
	}

	outcome := entity.IngestOutcomeUpdated
	if row.Inserted {
		outcome = entity.IngestOutcomeCreated
	}
	return toEntity(m), outcome, nil
}

func (r *repo) AttachToParent(ctx context.Context, childID, parentID string) error {
	values := dbpostgres.M{"parent_alert_id": parentID, "updated_at": time.Now().UTC()}
	if _, err := dbpostgres.Alerts(qm.Where("id = ?", childID)).UpdateAll(ctx, r.db.Querier(ctx), values); err != nil {
		return fmt.Errorf("attach alert to parent: %w", err)
	}
	return nil
}

func (r *repo) DetachFromParent(ctx context.Context, childID string) error {
	values := dbpostgres.M{"parent_alert_id": nil, "updated_at": time.Now().UTC()}
	if _, err := dbpostgres.Alerts(qm.Where("id = ?", childID)).UpdateAll(ctx, r.db.Querier(ctx), values); err != nil {
		return fmt.Errorf("detach alert from parent: %w", err)
	}
	return nil
}

func (r *repo) RollUpParent(ctx context.Context, parentID string, at time.Time) (entity.Alert, error) {
	if _, err := r.db.Querier(ctx).ExecContext(ctx, rollUpParentSQL, parentID, at); err != nil {
		return entity.Alert{}, fmt.Errorf("roll up group parent: %w", err)
	}
	m, err := dbpostgres.FindAlert(ctx, r.db.Querier(ctx), parentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Alert{}, entity.ErrAlertNotFound
		}
		return entity.Alert{}, fmt.Errorf("reload group parent: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) ListChildren(ctx context.Context, parentIDs []string) (map[string][]entity.AlertChild, error) {
	out := map[string][]entity.AlertChild{}
	if len(parentIDs) == 0 {
		return out, nil
	}
	rows, err := dbpostgres.Alerts(
		qm.WhereIn("parent_alert_id IN ?", anySlice(parentIDs)...),
		qm.OrderBy("last_seen_at DESC, id DESC"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list group children: %w", err)
	}
	for _, m := range rows {
		parent := m.ParentAlertID.String
		out[parent] = append(out[parent], entity.AlertChild{
			ID:         m.ID,
			Title:      m.Title,
			Count:      m.Count,
			LastSeenAt: m.LastSeenAt,
			Status:     entity.AlertStatus(m.Status),
			Severity:   entity.AlertSeverity(m.Severity),
		})
	}
	return out, nil
}

func (r *repo) ExpireStale(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		limit = entity.AlertListMaxPageSize
	}
	result, err := r.db.Querier(ctx).ExecContext(ctx, expireStaleSQL, now, limit)
	if err != nil {
		return 0, fmt.Errorf("expire stale alerts: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("expire stale alerts: %w", err)
	}
	return int(affected), nil
}

func (r *repo) CountsBySource(ctx context.Context, sourceIDs []string, since time.Time, buckets int) (map[string][]int, error) {
	out := make(map[string][]int, len(sourceIDs))
	for _, id := range sourceIDs {
		out[id] = make([]int, buckets)
	}
	if len(sourceIDs) == 0 || buckets <= 0 {
		return out, nil
	}

	width := entity.SourceVolumeWindow.Seconds() / float64(buckets)
	var rows []struct {
		SourceID string `boil:"source_id"`
		Bucket   int    `boil:"bucket"`
		Total    int    `boil:"total"`
	}
	err := queries.Raw(countsBySourceSQL, types.StringArray(sourceIDs), since, width, buckets-1).
		Bind(ctx, r.db.Querier(ctx), &rows)
	if err != nil {
		return nil, fmt.Errorf("count source events: %w", err)
	}
	for _, row := range rows {
		series, ok := out[row.SourceID]
		if !ok || row.Bucket < 0 || row.Bucket >= buckets {
			continue
		}
		series[row.Bucket] = row.Total
	}
	return out, nil
}

func (r *repo) ApplyRouting(ctx context.Context, alertID, policyRef, groupKey, silenceID string, suppressedAt time.Time) error {
	values := dbpostgres.M{
		"routed_policy_ref": policyRef,
		"group_key":         groupKey,
		"updated_at":        time.Now().UTC(),
	}
	if silenceID != "" {
		values["suppressed_by_silence_id"] = silenceID
		values["suppressed_at"] = suppressedAt
	}
	if _, err := dbpostgres.Alerts(qm.Where("id = ?", alertID)).UpdateAll(ctx, r.db.Querier(ctx), values); err != nil {
		return fmt.Errorf("apply alert routing: %w", err)
	}
	return nil
}
