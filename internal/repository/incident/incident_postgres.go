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

	events, err := dbpostgres.IncidentEvents(
		qm.Where("incident_id = ?", id),
		qm.OrderBy("at ASC, id ASC"),
		qm.Limit(entity.IncidentTimelineLimit),
	).All(ctx, exec)
	if err != nil {
		return entity.Incident{}, fmt.Errorf("list incident events: %w", err)
	}
	inc.Timeline = make([]entity.IncidentEvent, 0, len(events))
	for _, e := range events {
		inc.Timeline = append(inc.Timeline, entity.IncidentEvent{
			ID:         e.ID,
			IncidentID: e.IncidentID,
			At:         e.At,
			Kind:       e.Kind,
			Text:       e.Text,
			Actor:      e.Actor,
		})
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
	if _, err := dbpostgres.IncidentAlerts(
		qm.Where("workspace_id = ? AND incident_id = ? AND alert_id = ?", workspaceID, incidentID, alertID),
	).DeleteAll(ctx, r.db.Querier(ctx)); err != nil {
		return fmt.Errorf("unlink alert: %w", err)
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

func (r *repo) AppendEvent(ctx context.Context, event entity.IncidentEvent) error {
	m := &dbpostgres.IncidentEvent{
		IncidentID: event.IncidentID,
		At:         event.At,
		Kind:       event.Kind,
		Text:       event.Text,
		Actor:      event.Actor,
	}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("incident_id", "at", "kind", "text", "actor")); err != nil {
		return fmt.Errorf("append incident event: %w", err)
	}
	return nil
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
