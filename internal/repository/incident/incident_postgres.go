package incident

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

const incidentEventSavepoint = "incident_event_append"

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Incident {
	return &repo{db: db}
}

func anySlice(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}

func nullString(value string) null.String {
	if value == "" {
		return null.String{}
	}
	return null.StringFrom(value)
}

func customFieldsToJSON(fields map[string]string) (types.JSON, error) {
	if fields == nil {
		fields = map[string]string{}
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("marshal custom fields: %w", err)
	}
	return types.JSON(raw), nil
}

func customFieldsFromJSON(raw types.JSON) map[string]string {
	out := map[string]string{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func toEntity(m *dbpostgres.Incident) entity.Incident {
	inc := entity.Incident{
		ID:                m.ID,
		WorkspaceID:       m.WorkspaceID,
		Number:            m.Number,
		Name:              m.Name,
		Summary:           m.Summary,
		SeverityLevel:     m.SeverityLevel,
		Status:            entity.IncidentStatus(m.Status),
		LeadUserID:        m.LeadUserID.String,
		TeamID:            m.TeamID.String,
		CustomFields:      customFieldsFromJSON(m.CustomFields),
		ResolutionSummary: m.ResolutionSummary,
		DeclaredBy:        m.DeclaredBy.String,
		DeclaredAt:        m.DeclaredAt,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
	if m.ResolvedAt.Valid {
		inc.ResolvedAt = m.ResolvedAt.Time
	}
	return inc
}

func followupToEntity(m *dbpostgres.IncidentFollowup) entity.IncidentFollowup {
	f := entity.IncidentFollowup{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		IncidentID:  m.IncidentID,
		Title:       m.Title,
		OwnerUserID: m.OwnerUserID.String,
		Done:        m.Done,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.DueAt.Valid {
		f.DueAt = m.DueAt.Time
	}
	if m.DoneAt.Valid {
		f.DoneAt = m.DoneAt.Time
	}
	return f
}

func (r *repo) NextNumber(ctx context.Context, workspaceID string) (int, error) {
	var row struct {
		N int `boil:"n"`
	}
	err := queries.Raw(
		`SELECT COALESCE(MAX(number), 0) + 1 AS n FROM incidents WHERE workspace_id = $1`,
		workspaceID,
	).Bind(ctx, r.db.Querier(ctx), &row)
	if err != nil {
		return 0, fmt.Errorf("next incident number: %w", err)
	}
	return row.N, nil
}

func (r *repo) Create(ctx context.Context, in entity.Incident) (entity.Incident, error) {
	fields, err := customFieldsToJSON(in.CustomFields)
	if err != nil {
		return entity.Incident{}, err
	}
	m := &dbpostgres.Incident{
		WorkspaceID:   in.WorkspaceID,
		Number:        in.Number,
		Name:          in.Name,
		Summary:       in.Summary,
		SeverityLevel: in.SeverityLevel,
		Status:        string(entity.IncidentStatusDeclared),
		CustomFields:  fields,
		DeclaredAt:    in.DeclaredAt,
	}
	if in.LeadUserID != "" {
		m.LeadUserID = null.StringFrom(in.LeadUserID)
	}
	if in.TeamID != "" {
		m.TeamID = null.StringFrom(in.TeamID)
	}
	if in.DeclaredBy != "" {
		m.DeclaredBy = null.StringFrom(in.DeclaredBy)
	}
	cols := boil.Whitelist("workspace_id", "number", "name", "summary", "severity_level", "status", "lead_user_id", "team_id", "custom_fields", "declared_by", "declared_at")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.Incident{}, fmt.Errorf("create incident: %w", err)
		}
		return entity.Incident{}, fmt.Errorf("create incident: %w", err)
	}
	return r.GetByID(ctx, in.WorkspaceID, m.ID)
}

func (r *repo) GetByID(ctx context.Context, workspaceID, id string) (entity.Incident, error) {
	exec := r.db.Querier(ctx)
	m, err := dbpostgres.Incidents(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
	).One(ctx, exec)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Incident{}, entity.ErrIncidentNotFound
	}
	if err != nil {
		return entity.Incident{}, fmt.Errorf("get incident: %w", err)
	}
	inc := toEntity(m)

	services, err := r.servicesByIncident(ctx, exec, []string{id})
	if err != nil {
		return entity.Incident{}, err
	}
	inc.Services = services[id]

	alerts, err := r.alertsByIncident(ctx, exec, []string{id})
	if err != nil {
		return entity.Incident{}, err
	}
	inc.Alerts = alerts[id]

	relations, err := r.relationsByIncident(ctx, exec, workspaceID, id)
	if err != nil {
		return entity.Incident{}, err
	}
	inc.Related = relations

	followups, err := dbpostgres.IncidentFollowups(
		qm.Where("incident_id = ?", id),
		qm.OrderBy("done ASC, due_at ASC NULLS LAST, created_at ASC"),
	).All(ctx, exec)
	if err != nil {
		return entity.Incident{}, fmt.Errorf("list incident follow-ups: %w", err)
	}
	inc.Followups = make([]entity.IncidentFollowup, 0, len(followups))
	for _, f := range followups {
		inc.Followups = append(inc.Followups, followupToEntity(f))
	}

	return inc, nil
}

func (r *repo) servicesByIncident(ctx context.Context, exec boil.ContextExecutor, incidentIDs []string) (map[string][]entity.Service, error) {
	out := map[string][]entity.Service{}
	if len(incidentIDs) == 0 {
		return out, nil
	}
	links, err := dbpostgres.IncidentServices(
		qm.WhereIn("incident_id IN ?", anySlice(incidentIDs)...),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list incident services: %w", err)
	}
	if len(links) == 0 {
		return out, nil
	}
	ids := map[string]struct{}{}
	for _, l := range links {
		ids[l.ServiceID] = struct{}{}
	}
	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	services, err := dbpostgres.Services(
		qm.WhereIn("id IN ?", anySlice(idList)...),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("load services: %w", err)
	}
	byID := map[string]entity.Service{}
	for _, s := range services {
		byID[s.ID] = entity.Service{
			ID:          s.ID,
			WorkspaceID: s.WorkspaceID,
			Slug:        s.Slug,
			Name:        s.Name,
			TeamID:      s.TeamID.String,
			Description: s.Description,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		}
	}
	for _, l := range links {
		if s, ok := byID[l.ServiceID]; ok {
			out[l.IncidentID] = append(out[l.IncidentID], s)
		}
	}
	return out, nil
}

func (r *repo) alertsByIncident(ctx context.Context, exec boil.ContextExecutor, incidentIDs []string) (map[string][]entity.IncidentAlert, error) {
	out := map[string][]entity.IncidentAlert{}
	if len(incidentIDs) == 0 {
		return out, nil
	}
	links, err := dbpostgres.IncidentAlerts(
		qm.WhereIn("incident_id IN ?", anySlice(incidentIDs)...),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list incident alerts: %w", err)
	}
	if len(links) == 0 {
		return out, nil
	}
	ids := map[string]struct{}{}
	for _, l := range links {
		ids[l.AlertID] = struct{}{}
	}
	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	alerts, err := dbpostgres.Alerts(
		qm.Select("id", "title", "severity", "status"),
		qm.WhereIn("id IN ?", anySlice(idList)...),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("load alerts: %w", err)
	}
	byID := map[string]entity.IncidentAlert{}
	for _, a := range alerts {
		byID[a.ID] = entity.IncidentAlert{
			AlertID:  a.ID,
			Title:    a.Title,
			Severity: entity.AlertSeverity(a.Severity),
			Status:   entity.AlertStatus(a.Status),
		}
	}
	for _, l := range links {
		if a, ok := byID[l.AlertID]; ok {
			out[l.IncidentID] = append(out[l.IncidentID], a)
		}
	}
	return out, nil
}

func (r *repo) relationsByIncident(ctx context.Context, exec boil.ContextExecutor, workspaceID, incidentID string) ([]entity.IncidentRelation, error) {
	links, err := dbpostgres.IncidentRelations(
		qm.Where("workspace_id = ? AND incident_id = ?", workspaceID, incidentID),
		qm.OrderBy("created_at ASC"),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list incident relations: %w", err)
	}
	if len(links) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.RelatedIncidentID)
	}
	related, err := dbpostgres.Incidents(
		qm.Select("id", "number", "name", "status"),
		qm.WhereIn("id IN ?", anySlice(ids)...),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("load related incidents: %w", err)
	}
	byID := map[string]*dbpostgres.Incident{}
	for _, inc := range related {
		byID[inc.ID] = inc
	}
	out := make([]entity.IncidentRelation, 0, len(links))
	for _, l := range links {
		rel := entity.IncidentRelation{
			ID:        l.ID,
			Kind:      entity.IncidentRelationKind(l.Kind),
			RelatedID: l.RelatedIncidentID,
		}
		if inc, ok := byID[l.RelatedIncidentID]; ok {
			rel.RelatedNumber = inc.Number
			rel.RelatedName = inc.Name
			rel.RelatedStatus = entity.IncidentStatus(inc.Status)
		}
		out = append(out, rel)
	}
	return out, nil
}

func (r *repo) List(ctx context.Context, workspaceID string, filter entity.IncidentFilter) (entity.IncidentPage, error) {
	limit := filter.Limit
	if limit <= 0 || limit > entity.IncidentListMaxPageSize {
		limit = entity.IncidentListMaxPageSize
	}
	exec := r.db.Querier(ctx)
	mods := []qm.QueryMod{qm.Where("workspace_id = ?", workspaceID)}
	if len(filter.Statuses) > 0 {
		mods = append(mods, qm.WhereIn("status IN ?", anySlice(statusStrings(filter.Statuses))...))
	}
	if filter.ActiveOnly {
		mods = append(mods, qm.Where("status <> ?", string(entity.IncidentStatusResolved)))
	}
	if len(filter.Severities) > 0 {
		mods = append(mods, qm.WhereIn("severity_level IN ?", anySlice(filter.Severities)...))
	}
	if len(filter.TeamIDs) > 0 {
		mods = append(mods, qm.WhereIn("team_id IN ?", anySlice(filter.TeamIDs)...))
	}
	if len(filter.ServiceIDs) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(filter.ServiceIDs)), ",")
		mods = append(mods, qm.Where(
			"EXISTS (SELECT 1 FROM incident_services isv WHERE isv.incident_id = incidents.id AND isv.service_id IN ("+placeholders+"))",
			anySlice(filter.ServiceIDs)...,
		))
	}
	if !filter.Since.IsZero() {
		mods = append(mods, qm.Where("declared_at >= ?", filter.Since))
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		mods = append(mods, qm.Where("(lower(name) LIKE ? OR lower(summary) LIKE ?)", like, like))
	}
	if cursorAt, cursorID, ok := decodeCursor(filter.Cursor); ok {
		mods = append(mods, qm.Where("(declared_at, id) < (?, ?)", cursorAt, cursorID))
	}
	mods = append(mods, qm.OrderBy("declared_at DESC, id DESC"), qm.Limit(limit+1))

	rows, err := dbpostgres.Incidents(mods...).All(ctx, exec)
	if err != nil {
		return entity.IncidentPage{}, fmt.Errorf("list incidents: %w", err)
	}

	page := entity.IncidentPage{}
	if len(rows) > limit {
		last := rows[limit-1]
		page.NextCursor = encodeCursor(last.DeclaredAt, last.ID)
		rows = rows[:limit]
	}
	ids := make([]string, 0, len(rows))
	for _, m := range rows {
		ids = append(ids, m.ID)
	}
	services, err := r.servicesByIncident(ctx, exec, ids)
	if err != nil {
		return entity.IncidentPage{}, err
	}
	page.Incidents = make([]entity.Incident, 0, len(rows))
	for _, m := range rows {
		inc := toEntity(m)
		inc.Services = services[m.ID]
		page.Incidents = append(page.Incidents, inc)
	}
	return page, nil
}

func (r *repo) Patch(ctx context.Context, workspaceID, id string, patch entity.IncidentPatch) (bool, error) {
	values := dbpostgres.M{"updated_at": time.Now().UTC()}
	if patch.Name != nil {
		values["name"] = *patch.Name
	}
	if patch.Summary != nil {
		values["summary"] = *patch.Summary
	}
	if patch.SeverityLevel != nil {
		values["severity_level"] = *patch.SeverityLevel
	}
	if patch.LeadUserID != nil {
		if *patch.LeadUserID == "" {
			values["lead_user_id"] = nil
		} else {
			values["lead_user_id"] = *patch.LeadUserID
		}
	}
	if patch.TeamID != nil {
		if *patch.TeamID == "" {
			values["team_id"] = nil
		} else {
			values["team_id"] = *patch.TeamID
		}
	}
	affected, err := dbpostgres.Incidents(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("patch incident: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) SetStatus(ctx context.Context, workspaceID, id string, from, to entity.IncidentStatus, at time.Time, resolution string) (bool, error) {
	values := dbpostgres.M{"status": string(to), "updated_at": at}
	if to == entity.IncidentStatusResolved {
		values["resolved_at"] = at
		values["resolution_summary"] = resolution
	} else {
		values["resolved_at"] = nil
	}
	affected, err := dbpostgres.Incidents(
		qm.Where("workspace_id = ? AND id = ? AND status = ?", workspaceID, id, string(from)),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("set incident status: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) Reopen(ctx context.Context, workspaceID, id string, to entity.IncidentStatus, at time.Time) (bool, error) {
	values := dbpostgres.M{"status": string(to), "resolved_at": nil, "updated_at": at}
	affected, err := dbpostgres.Incidents(
		qm.Where("workspace_id = ? AND id = ? AND status = ?", workspaceID, id, string(entity.IncidentStatusResolved)),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("reopen incident: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) SetCustomFields(ctx context.Context, workspaceID, id string, fields map[string]string) (bool, error) {
	raw, err := customFieldsToJSON(fields)
	if err != nil {
		return false, err
	}
	values := dbpostgres.M{"custom_fields": raw, "updated_at": time.Now().UTC()}
	affected, err := dbpostgres.Incidents(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("set custom fields: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) ReplaceServices(ctx context.Context, workspaceID, incidentID string, serviceIDs []string) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.IncidentServices(
		qm.Where("incident_id = ?", incidentID),
	).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("clear incident services: %w", err)
	}
	seen := map[string]struct{}{}
	for _, sid := range serviceIDs {
		if _, ok := seen[sid]; ok {
			continue
		}
		seen[sid] = struct{}{}
		m := &dbpostgres.IncidentService{IncidentID: incidentID, ServiceID: sid, WorkspaceID: workspaceID}
		if err := m.Insert(ctx, exec, boil.Whitelist("incident_id", "service_id", "workspace_id")); err != nil {
			return fmt.Errorf("link incident service: %w", err)
		}
	}
	return nil
}

func (r *repo) LinkAlert(ctx context.Context, workspaceID, incidentID, alertID string) error {
	m := &dbpostgres.IncidentAlert{IncidentID: incidentID, AlertID: alertID, WorkspaceID: workspaceID}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("incident_id", "alert_id", "workspace_id")); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return nil
		}
		return fmt.Errorf("link alert: %w", err)
	}
	return nil
}

func (r *repo) UnlinkAlert(ctx context.Context, workspaceID, incidentID, alertID string) error {
	affected, err := dbpostgres.IncidentAlerts(
		qm.Where("workspace_id = ? AND incident_id = ? AND alert_id = ?", workspaceID, incidentID, alertID),
	).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("unlink alert: %w", err)
	}
	if affected == 0 {
		return entity.ErrAlertNotFound
	}
	return nil
}

func (r *repo) LinkedAlertIDs(ctx context.Context, incidentID string) ([]string, error) {
	rows, err := dbpostgres.IncidentAlerts(
		qm.Select("alert_id"),
		qm.Where("incident_id = ?", incidentID),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list linked alerts: %w", err)
	}
	out := make([]string, 0, len(rows))
	for _, m := range rows {
		out = append(out, m.AlertID)
	}
	return out, nil
}

func (r *repo) Relate(ctx context.Context, workspaceID, incidentID, relatedID string, kind entity.IncidentRelationKind) (entity.IncidentRelation, error) {
	m := &dbpostgres.IncidentRelation{
		WorkspaceID:       workspaceID,
		IncidentID:        incidentID,
		RelatedIncidentID: relatedID,
		Kind:              string(kind),
	}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("workspace_id", "incident_id", "related_incident_id", "kind")); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.IncidentRelation{}, nil
		}
		return entity.IncidentRelation{}, fmt.Errorf("relate incident: %w", err)
	}
	return entity.IncidentRelation{ID: m.ID, Kind: kind, RelatedID: relatedID}, nil
}

func (r *repo) Unrelate(ctx context.Context, workspaceID, incidentID, relationID string) error {
	affected, err := dbpostgres.IncidentRelations(
		qm.Where("workspace_id = ? AND incident_id = ? AND id = ?", workspaceID, incidentID, relationID),
	).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("unrelate incident: %w", err)
	}
	if affected == 0 {
		return entity.ErrIncidentNotFound
	}
	return nil
}

func (r *repo) AppendEvent(ctx context.Context, event entity.IncidentEvent) (entity.IncidentEvent, error) {
	if event.IdempotencyKey != "" {
		existing, err := r.eventByKey(ctx, event.IncidentID, event.IdempotencyKey)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return entity.IncidentEvent{}, err
		}
		if err == nil {
			return existing, nil
		}
	}
	m := &dbpostgres.IncidentEvent{
		IncidentID:     event.IncidentID,
		WorkspaceID:    event.WorkspaceID,
		At:             event.At,
		Kind:           string(event.Kind),
		Category:       string(event.Category),
		Source:         string(event.Source),
		Text:           event.Text,
		Actor:          event.Actor,
		Retroactive:    event.Retroactive,
		ActorUserID:    nullString(event.ActorUserID),
		IdempotencyKey: event.IdempotencyKey,
	}
	err := r.db.WithSavepoint(ctx, incidentEventSavepoint, func(ctx context.Context) error {
		return m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist(
			"incident_id", "workspace_id", "at", "kind", "category", "source",
			"text", "actor", "retroactive", "actor_user_id", "idempotency_key",
		))
	})
	if err != nil {
		if _, ok := postgres.UniqueViolation(err); ok && event.IdempotencyKey != "" {
			existing, findErr := r.eventByKey(ctx, event.IncidentID, event.IdempotencyKey)
			if findErr != nil {
				return entity.IncidentEvent{}, findErr
			}
			return existing, nil
		}
		return entity.IncidentEvent{}, fmt.Errorf("append incident event: %w", err)
	}
	return eventToEntity(m), nil
}

func (r *repo) eventByKey(ctx context.Context, incidentID, idempotencyKey string) (entity.IncidentEvent, error) {
	m, err := dbpostgres.IncidentEvents(
		qm.Where("incident_id = ? AND idempotency_key = ?", incidentID, idempotencyKey),
	).One(ctx, r.db.Querier(ctx))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.IncidentEvent{}, err
	}
	if err != nil {
		return entity.IncidentEvent{}, fmt.Errorf("load replayed incident event: %w", err)
	}
	return eventToEntity(m), nil
}

func (r *repo) ListEvents(ctx context.Context, workspaceID, incidentID string, categories []entity.IncidentEventCategory, after entity.TimelineCursor, limit int) ([]entity.IncidentEvent, error) {
	if limit <= 0 || limit > entity.TimelineFetchLimit {
		limit = entity.TimelineFetchLimit
	}
	mods := []qm.QueryMod{qm.Where("workspace_id = ? AND incident_id = ?", workspaceID, incidentID)}
	if len(categories) > 0 {
		raw := make([]string, len(categories))
		for i, c := range categories {
			raw[i] = string(c)
		}
		mods = append(mods, qm.WhereIn("category IN ?", anySlice(raw)...))
	}
	if !after.Zero() {
		mods = append(mods, qm.Where("(at, id) > (?, ?)", after.At, after.ID))
	}
	mods = append(mods, qm.OrderBy("at, id"), qm.Limit(limit))

	exec := r.db.Querier(ctx)
	rows, err := dbpostgres.IncidentEvents(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list incident events: %w", err)
	}
	out := make([]entity.IncidentEvent, 0, len(rows))
	ids := make([]string, 0, len(rows))
	for _, m := range rows {
		out = append(out, eventToEntity(m))
		ids = append(ids, m.ID)
	}
	attachments, err := r.attachmentsByEvent(ctx, exec, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Attachments = attachments[out[i].ID]
	}
	return out, nil
}

func (r *repo) GetEvent(ctx context.Context, workspaceID, eventID string) (entity.IncidentEvent, error) {
	exec := r.db.Querier(ctx)
	m, err := dbpostgres.IncidentEvents(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, eventID),
	).One(ctx, exec)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.IncidentEvent{}, entity.ErrTimelineEntryNotFound
	}
	if err != nil {
		return entity.IncidentEvent{}, fmt.Errorf("get incident event: %w", err)
	}
	event := eventToEntity(m)
	attachments, err := r.attachmentsByEvent(ctx, exec, []string{m.ID})
	if err != nil {
		return entity.IncidentEvent{}, err
	}
	event.Attachments = attachments[m.ID]
	return event, nil
}

func (r *repo) GetEventForUpdate(ctx context.Context, workspaceID, eventID string) (entity.IncidentEvent, error) {
	m, err := dbpostgres.IncidentEvents(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, eventID),
		qm.For("UPDATE"),
	).One(ctx, r.db.Querier(ctx))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.IncidentEvent{}, entity.ErrTimelineEntryNotFound
	}
	if err != nil {
		return entity.IncidentEvent{}, fmt.Errorf("lock incident event: %w", err)
	}
	return eventToEntity(m), nil
}

func (r *repo) UpdateEvent(ctx context.Context, workspaceID, eventID string, edit entity.TimelineEdit, editedAt time.Time, editorUserID string) error {
	affected, err := dbpostgres.IncidentEvents(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, eventID),
	).UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{
		"text":      edit.Text,
		"category":  string(edit.Category),
		"edited_at": editedAt,
		"edited_by": nullString(editorUserID),
	})
	if err != nil {
		return fmt.Errorf("update incident event: %w", err)
	}
	if affected == 0 {
		return entity.ErrTimelineEntryNotFound
	}
	return nil
}

func (r *repo) AppendRevision(ctx context.Context, revision entity.IncidentEventRevision) error {
	m := &dbpostgres.IncidentEventRevision{
		EventID:      revision.EventID,
		WorkspaceID:  revision.WorkspaceID,
		At:           revision.At,
		EditorUserID: nullString(revision.EditorUserID),
		EditorLabel:  revision.EditorLabel,
		Text:         revision.Text,
		Category:     string(revision.Category),
	}
	columns := boil.Whitelist("event_id", "workspace_id", "at", "editor_user_id", "editor_label", "text", "category")
	if err := m.Insert(ctx, r.db.Querier(ctx), columns); err != nil {
		return fmt.Errorf("append incident event revision: %w", err)
	}
	return nil
}

func (r *repo) ListRevisions(ctx context.Context, workspaceID, eventID string) ([]entity.IncidentEventRevision, error) {
	rows, err := dbpostgres.IncidentEventRevisions(
		qm.Where("workspace_id = ? AND event_id = ?", workspaceID, eventID),
		qm.OrderBy("at, id"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list incident event revisions: %w", err)
	}
	out := make([]entity.IncidentEventRevision, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.IncidentEventRevision{
			ID:           m.ID,
			EventID:      m.EventID,
			WorkspaceID:  m.WorkspaceID,
			At:           m.At,
			EditorUserID: m.EditorUserID.String,
			EditorLabel:  m.EditorLabel,
			Text:         m.Text,
			Category:     entity.IncidentEventCategory(m.Category),
		})
	}
	return out, nil
}

func (r *repo) AddAttachment(ctx context.Context, attachment entity.IncidentEventAttachment) (entity.IncidentEventAttachment, error) {
	m := &dbpostgres.IncidentEventAttachment{
		EventID:     attachment.EventID,
		WorkspaceID: attachment.WorkspaceID,
		Kind:        string(attachment.Kind),
		Label:       attachment.Label,
		URL:         attachment.URL,
		Body:        attachment.Body,
		ObjectKey:   attachment.ObjectKey,
		ContentType: attachment.ContentType,
		SizeBytes:   attachment.SizeBytes,
		CreatedBy:   nullString(attachment.CreatedBy),
	}
	columns := boil.Whitelist("event_id", "workspace_id", "kind", "label", "url", "body", "object_key", "content_type", "size_bytes", "created_by")
	if err := m.Insert(ctx, r.db.Querier(ctx), columns); err != nil {
		return entity.IncidentEventAttachment{}, fmt.Errorf("add incident event attachment: %w", err)
	}
	return attachmentToEntity(m), nil
}

func (r *repo) GetAttachment(ctx context.Context, workspaceID, attachmentID string) (entity.IncidentEventAttachment, error) {
	m, err := dbpostgres.IncidentEventAttachments(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, attachmentID),
	).One(ctx, r.db.Querier(ctx))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.IncidentEventAttachment{}, entity.ErrAttachmentNotFound
	}
	if err != nil {
		return entity.IncidentEventAttachment{}, fmt.Errorf("get incident event attachment: %w", err)
	}
	return attachmentToEntity(m), nil
}

func (r *repo) RemoveAttachment(ctx context.Context, workspaceID, attachmentID string) error {
	affected, err := dbpostgres.IncidentEventAttachments(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, attachmentID),
	).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("remove incident event attachment: %w", err)
	}
	if affected == 0 {
		return entity.ErrAttachmentNotFound
	}
	return nil
}

func (r *repo) CountAttachments(ctx context.Context, workspaceID, eventID string) (int, error) {
	total, err := dbpostgres.IncidentEventAttachments(
		qm.Where("workspace_id = ? AND event_id = ?", workspaceID, eventID),
	).Count(ctx, r.db.Querier(ctx))
	if err != nil {
		return 0, fmt.Errorf("count incident event attachments: %w", err)
	}
	return int(total), nil
}

func (r *repo) attachmentsByEvent(ctx context.Context, exec boil.ContextExecutor, eventIDs []string) (map[string][]entity.IncidentEventAttachment, error) {
	out := map[string][]entity.IncidentEventAttachment{}
	if len(eventIDs) == 0 {
		return out, nil
	}
	rows, err := dbpostgres.IncidentEventAttachments(
		qm.WhereIn("event_id IN ?", anySlice(eventIDs)...),
		qm.OrderBy("created_at, id"),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list incident event attachments: %w", err)
	}
	for _, m := range rows {
		out[m.EventID] = append(out[m.EventID], attachmentToEntity(m))
	}
	return out, nil
}

func eventToEntity(m *dbpostgres.IncidentEvent) entity.IncidentEvent {
	return entity.IncidentEvent{
		ID:             m.ID,
		IncidentID:     m.IncidentID,
		WorkspaceID:    m.WorkspaceID,
		At:             m.At,
		Kind:           entity.IncidentEventKind(m.Kind),
		Category:       entity.IncidentEventCategory(m.Category),
		Source:         entity.IncidentEventSource(m.Source),
		Text:           m.Text,
		Actor:          m.Actor,
		ActorUserID:    m.ActorUserID.String,
		Retroactive:    m.Retroactive,
		EditedAt:       m.EditedAt.Time,
		EditedBy:       m.EditedBy.String,
		IdempotencyKey: m.IdempotencyKey,
	}
}

func attachmentToEntity(m *dbpostgres.IncidentEventAttachment) entity.IncidentEventAttachment {
	return entity.IncidentEventAttachment{
		ID:          m.ID,
		EventID:     m.EventID,
		WorkspaceID: m.WorkspaceID,
		Kind:        entity.AttachmentKind(m.Kind),
		Label:       m.Label,
		URL:         m.URL,
		Body:        m.Body,
		ObjectKey:   m.ObjectKey,
		ContentType: m.ContentType,
		SizeBytes:   m.SizeBytes,
		CreatedAt:   m.CreatedAt,
		CreatedBy:   m.CreatedBy.String,
	}
}

func (r *repo) AddFollowup(ctx context.Context, f entity.IncidentFollowup) (entity.IncidentFollowup, error) {
	exec := r.db.Querier(ctx)
	m := &dbpostgres.IncidentFollowup{
		WorkspaceID: f.WorkspaceID,
		IncidentID:  f.IncidentID,
		Title:       f.Title,
	}
	if f.OwnerUserID != "" {
		m.OwnerUserID = null.StringFrom(f.OwnerUserID)
	}
	if !f.DueAt.IsZero() {
		m.DueAt = null.TimeFrom(f.DueAt)
	}
	if err := m.Insert(ctx, exec, boil.Whitelist("workspace_id", "incident_id", "title", "owner_user_id", "due_at")); err != nil {
		return entity.IncidentFollowup{}, fmt.Errorf("add follow-up: %w", err)
	}
	reloaded, err := dbpostgres.IncidentFollowups(qm.Where("id = ?", m.ID)).One(ctx, exec)
	if err != nil {
		return entity.IncidentFollowup{}, fmt.Errorf("load follow-up: %w", err)
	}
	return followupToEntity(reloaded), nil
}

func (r *repo) SetFollowupDone(ctx context.Context, workspaceID, id string, done bool, at time.Time) (entity.IncidentFollowup, error) {
	exec := r.db.Querier(ctx)
	values := dbpostgres.M{"done": done, "updated_at": at}
	if done {
		values["done_at"] = at
	} else {
		values["done_at"] = nil
	}
	affected, err := dbpostgres.IncidentFollowups(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
	).UpdateAll(ctx, exec, values)
	if err != nil {
		return entity.IncidentFollowup{}, fmt.Errorf("update follow-up: %w", err)
	}
	if affected == 0 {
		return entity.IncidentFollowup{}, entity.ErrFollowupNotFound
	}
	reloaded, err := dbpostgres.IncidentFollowups(qm.Where("id = ?", id)).One(ctx, exec)
	if err != nil {
		return entity.IncidentFollowup{}, fmt.Errorf("load follow-up: %w", err)
	}
	return followupToEntity(reloaded), nil
}

func (r *repo) ListFollowups(ctx context.Context, workspaceID string) ([]entity.IncidentFollowup, error) {
	rows, err := dbpostgres.IncidentFollowups(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("done ASC, due_at ASC NULLS LAST, created_at ASC"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list follow-ups: %w", err)
	}
	out := make([]entity.IncidentFollowup, 0, len(rows))
	for _, m := range rows {
		out = append(out, followupToEntity(m))
	}
	return out, nil
}

func statusStrings(in []entity.IncidentStatus) []string {
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
