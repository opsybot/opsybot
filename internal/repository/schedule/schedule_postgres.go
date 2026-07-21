package schedule

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func New(db postgres.Client) repository.Schedule {
	return &repo{db: db}
}

func anySlice(ids []string) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

func clampInterval(days int) int {
	if days < entity.LayerMinIntervalDays {
		return entity.LayerMinIntervalDays
	}
	if days > entity.LayerMaxIntervalDays {
		return entity.LayerMaxIntervalDays
	}
	return days
}

func scheduleBase(m *dbpostgres.Schedule, teamSlug string) entity.Schedule {
	return entity.Schedule{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		TeamID:      m.TeamID,
		TeamSlug:    teamSlug,
		Slug:        m.Slug,
		Timezone:    m.Timezone,
		FeedToken:   m.FeedToken,
		Paused:      m.PausedAt.Valid,
		Archived:    m.ArchivedAt.Valid,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (r *repo) getMany(ctx context.Context, mods ...qm.QueryMod) ([]entity.Schedule, error) {
	exec := r.db.Querier(ctx)
	rows, err := dbpostgres.Schedules(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	if len(rows) == 0 {
		return []entity.Schedule{}, nil
	}

	scheduleIDs := make([]string, len(rows))
	teamIDs := make([]string, 0, len(rows))
	for i, m := range rows {
		scheduleIDs[i] = m.ID
		teamIDs = append(teamIDs, m.TeamID)
	}

	teamRows, err := dbpostgres.Teams(qm.WhereIn("id in ?", anySlice(teamIDs)...)).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list schedule teams: %w", err)
	}
	teamSlug := make(map[string]string, len(teamRows))
	for _, t := range teamRows {
		teamSlug[t.ID] = t.Slug
	}

	layerRows, err := dbpostgres.ScheduleLayers(
		qm.WhereIn("schedule_id in ?", anySlice(scheduleIDs)...),
		qm.OrderBy("schedule_id, position"),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list schedule layers: %w", err)
	}

	layerIDs := make([]string, len(layerRows))
	for i, l := range layerRows {
		layerIDs[i] = l.ID
	}

	partsByLayer := map[string][]string{}
	restsByLayer := map[string][]entity.Restriction{}
	if len(layerIDs) > 0 {
		partRows, err := dbpostgres.ScheduleLayerParticipants(
			qm.WhereIn("layer_id in ?", anySlice(layerIDs)...),
			qm.OrderBy("layer_id, position"),
		).All(ctx, exec)
		if err != nil {
			return nil, fmt.Errorf("list schedule participants: %w", err)
		}
		for _, p := range partRows {
			partsByLayer[p.LayerID] = append(partsByLayer[p.LayerID], p.UserID)
		}

		restRows, err := dbpostgres.ScheduleLayerRestrictions(
			qm.WhereIn("layer_id in ?", anySlice(layerIDs)...),
			qm.OrderBy("layer_id, weekday, start_minute"),
		).All(ctx, exec)
		if err != nil {
			return nil, fmt.Errorf("list schedule restrictions: %w", err)
		}
		for _, rr := range restRows {
			restsByLayer[rr.LayerID] = append(restsByLayer[rr.LayerID], entity.Restriction{
				Weekday: rr.Weekday, StartMinute: rr.StartMinute, EndMinute: rr.EndMinute,
			})
		}
	}

	layersBySchedule := map[string][]entity.Layer{}
	for _, l := range layerRows {
		layersBySchedule[l.ScheduleID] = append(layersBySchedule[l.ScheduleID], entity.Layer{
			ID:           l.ID,
			Participants: partsByLayer[l.ID],
			Rotation:     entity.Rotation(l.Rotation),
			IntervalDays: l.IntervalDays,
			HandoverHour: l.HandoverHour,
			StartsOn:     l.StartsOn,
			Restrictions: restsByLayer[l.ID],
		})
	}

	overrideRows, err := dbpostgres.ScheduleOverrides(
		qm.WhereIn("schedule_id in ?", anySlice(scheduleIDs)...),
		qm.OrderBy("created_at, id"),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list schedule overrides: %w", err)
	}
	overridesBySchedule := map[string][]entity.Override{}
	for _, o := range overrideRows {
		overridesBySchedule[o.ScheduleID] = append(overridesBySchedule[o.ScheduleID], overrideToEntity(o))
	}

	out := make([]entity.Schedule, 0, len(rows))
	for _, m := range rows {
		s := scheduleBase(m, teamSlug[m.TeamID])
		s.Layers = layersBySchedule[m.ID]
		s.Overrides = overridesBySchedule[m.ID]
		out = append(out, s)
	}
	return out, nil
}

func overrideToEntity(o *dbpostgres.ScheduleOverride) entity.Override {
	return entity.Override{
		ID:              o.ID,
		UserID:          o.UserID,
		StartsAt:        o.StartsAt,
		EndsAt:          o.EndsAt,
		Reason:          o.Reason,
		CreatedByUserID: o.CreatedByUserID.String,
		CreatedAt:       o.CreatedAt,
	}
}

func (r *repo) GetBySlug(ctx context.Context, workspaceID, slug string) (entity.Schedule, error) {
	out, err := r.getMany(ctx, qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug))
	if err != nil {
		return entity.Schedule{}, err
	}
	if len(out) == 0 {
		return entity.Schedule{}, entity.ErrScheduleNotFound
	}
	return out[0], nil
}

func (r *repo) GetByFeedToken(ctx context.Context, feedToken string) (entity.Schedule, error) {
	out, err := r.getMany(ctx, qm.Where("feed_token = ?", feedToken))
	if err != nil {
		return entity.Schedule{}, err
	}
	if len(out) == 0 {
		return entity.Schedule{}, entity.ErrScheduleNotFound
	}
	return out[0], nil
}

func (r *repo) ListByWorkspace(ctx context.Context, workspaceID string, includeArchived bool) ([]entity.Schedule, error) {
	mods := []qm.QueryMod{qm.Where("workspace_id = ?", workspaceID)}
	if !includeArchived {
		mods = append(mods, qm.Where("archived_at IS NULL"))
	}
	mods = append(mods, qm.OrderBy("slug"))
	return r.getMany(ctx, mods...)
}

func (r *repo) ListActive(ctx context.Context, workspaceID string) ([]entity.Schedule, error) {
	return r.getMany(ctx,
		qm.Where("workspace_id = ? AND archived_at IS NULL AND paused_at IS NULL", workspaceID),
		qm.OrderBy("slug"),
	)
}

func (r *repo) Create(ctx context.Context, s entity.Schedule) (entity.Schedule, error) {
	exec := r.db.Querier(ctx)
	m := &dbpostgres.Schedule{
		WorkspaceID: s.WorkspaceID,
		TeamID:      s.TeamID,
		Slug:        s.Slug,
		Timezone:    s.Timezone,
		FeedToken:   s.FeedToken,
	}
	if err := m.Insert(ctx, exec, boil.Whitelist("workspace_id", "team_id", "slug", "timezone", "feed_token")); err != nil {
		if name, ok := postgres.UniqueViolation(err); ok && strings.Contains(name, "slug") {
			return entity.Schedule{}, entity.ErrScheduleSlugTaken
		}
		return entity.Schedule{}, fmt.Errorf("create schedule: %w", err)
	}
	if err := r.insertLayers(ctx, exec, m.ID, s.WorkspaceID, s.Layers); err != nil {
		return entity.Schedule{}, err
	}
	return r.GetBySlug(ctx, s.WorkspaceID, s.Slug)
}

func (r *repo) Update(ctx context.Context, workspaceID, slug string, s entity.Schedule) (entity.Schedule, error) {
	exec := r.db.Querier(ctx)
	current, err := dbpostgres.Schedules(qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug)).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Schedule{}, entity.ErrScheduleNotFound
		}
		return entity.Schedule{}, fmt.Errorf("get schedule: %w", err)
	}
	if _, err := dbpostgres.Schedules(qm.Where("id = ?", current.ID)).UpdateAll(ctx, exec, dbpostgres.M{
		"slug":       s.Slug,
		"team_id":    s.TeamID,
		"timezone":   s.Timezone,
		"updated_at": time.Now(),
	}); err != nil {
		if name, ok := postgres.UniqueViolation(err); ok && strings.Contains(name, "slug") {
			return entity.Schedule{}, entity.ErrScheduleSlugTaken
		}
		return entity.Schedule{}, fmt.Errorf("update schedule: %w", err)
	}
	if _, err := dbpostgres.ScheduleLayers(qm.Where("schedule_id = ?", current.ID)).DeleteAll(ctx, exec); err != nil {
		return entity.Schedule{}, fmt.Errorf("clear schedule layers: %w", err)
	}
	if err := r.insertLayers(ctx, exec, current.ID, workspaceID, s.Layers); err != nil {
		return entity.Schedule{}, err
	}
	return r.GetBySlug(ctx, workspaceID, s.Slug)
}

func (r *repo) insertLayers(ctx context.Context, exec boil.ContextExecutor, scheduleID, workspaceID string, layers []entity.Layer) error {
	for i, layer := range layers {
		lm := &dbpostgres.ScheduleLayer{
			ScheduleID:   scheduleID,
			WorkspaceID:  workspaceID,
			Position:     i,
			Rotation:     string(layer.Rotation),
			IntervalDays: clampInterval(layer.IntervalDays),
			HandoverHour: layer.HandoverHour,
			StartsOn:     layer.StartsOn,
		}
		if err := lm.Insert(ctx, exec, boil.Whitelist("schedule_id", "workspace_id", "position", "rotation", "interval_days", "handover_hour", "starts_on")); err != nil {
			return fmt.Errorf("create schedule layer: %w", err)
		}
		for j, userID := range layer.Participants {
			pm := &dbpostgres.ScheduleLayerParticipant{LayerID: lm.ID, WorkspaceID: workspaceID, UserID: userID, Position: j}
			if err := pm.Insert(ctx, exec, boil.Whitelist("layer_id", "workspace_id", "user_id", "position")); err != nil {
				return fmt.Errorf("create schedule participant: %w", err)
			}
		}
		for _, rr := range layer.Restrictions {
			rm := &dbpostgres.ScheduleLayerRestriction{LayerID: lm.ID, Weekday: rr.Weekday, StartMinute: rr.StartMinute, EndMinute: rr.EndMinute}
			if err := rm.Insert(ctx, exec, boil.Whitelist("layer_id", "weekday", "start_minute", "end_minute")); err != nil {
				return fmt.Errorf("create schedule restriction: %w", err)
			}
		}
	}
	return nil
}

func (r *repo) setTimestamp(ctx context.Context, workspaceID, slug, column string, set bool) (entity.Schedule, error) {
	var value any
	if set {
		value = time.Now()
	}
	n, err := dbpostgres.Schedules(qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug)).
		UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{column: value, "updated_at": time.Now()})
	if err != nil {
		return entity.Schedule{}, fmt.Errorf("update schedule %s: %w", column, err)
	}
	if n == 0 {
		return entity.Schedule{}, entity.ErrScheduleNotFound
	}
	return r.GetBySlug(ctx, workspaceID, slug)
}

func (r *repo) SetArchived(ctx context.Context, workspaceID, slug string, archived bool) (entity.Schedule, error) {
	return r.setTimestamp(ctx, workspaceID, slug, "archived_at", archived)
}

func (r *repo) SetPaused(ctx context.Context, workspaceID, slug string, paused bool) (entity.Schedule, error) {
	return r.setTimestamp(ctx, workspaceID, slug, "paused_at", paused)
}

func (r *repo) Delete(ctx context.Context, workspaceID, slug string) error {
	n, err := dbpostgres.Schedules(qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug)).
		DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	if n == 0 {
		return entity.ErrScheduleNotFound
	}
	return nil
}

func (r *repo) SlugExists(ctx context.Context, workspaceID, slug string) (bool, error) {
	exists, err := dbpostgres.Schedules(qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug)).Exists(ctx, r.db.Querier(ctx))
	if err != nil {
		return false, fmt.Errorf("schedule slug exists: %w", err)
	}
	return exists, nil
}

func (r *repo) AddOverride(ctx context.Context, workspaceID, scheduleID string, o entity.Override) (entity.Override, error) {
	m := &dbpostgres.ScheduleOverride{
		ScheduleID:      scheduleID,
		WorkspaceID:     workspaceID,
		UserID:          o.UserID,
		StartsAt:        o.StartsAt,
		EndsAt:          o.EndsAt,
		Reason:          o.Reason,
		CreatedByUserID: null.NewString(o.CreatedByUserID, o.CreatedByUserID != ""),
	}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("schedule_id", "workspace_id", "user_id", "starts_at", "ends_at", "reason", "created_by_user_id")); err != nil {
		return entity.Override{}, fmt.Errorf("create schedule override: %w", err)
	}
	return overrideToEntity(m), nil
}

func (r *repo) ListReferencesByUser(ctx context.Context, workspaceID, userID string) ([]entity.MemberReference, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT DISTINCT s.id, s.slug
		 FROM schedule_layer_participants p
		 JOIN schedule_layers l ON l.id = p.layer_id
		 JOIN schedules s ON s.id = l.schedule_id
		 WHERE p.workspace_id = $1 AND p.user_id = $2 AND s.archived_at IS NULL
		 ORDER BY s.slug`, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("list schedule references: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []entity.MemberReference
	for rows.Next() {
		var id, slug string
		if err := rows.Scan(&id, &slug); err != nil {
			return nil, fmt.Errorf("scan schedule reference: %w", err)
		}
		out = append(out, entity.MemberReference{
			ID:     id + ":" + userID,
			Kind:   entity.ReferenceKindSchedule,
			Icon:   "calendar-clock",
			Label:  slug,
			Detail: "On-call schedule",
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schedule references: %w", err)
	}
	return out, nil
}

func (r *repo) Reassign(ctx context.Context, workspaceID, scheduleID, fromUserID, toUserID string) error {
	exec := r.db.Querier(ctx)
	if _, err := exec.ExecContext(ctx,
		`DELETE FROM schedule_layer_participants p
		 USING schedule_layers l
		 WHERE p.layer_id = l.id AND l.schedule_id = $1 AND p.workspace_id = $2 AND p.user_id = $3
		   AND EXISTS (SELECT 1 FROM schedule_layer_participants q WHERE q.layer_id = p.layer_id AND q.user_id = $4)`,
		scheduleID, workspaceID, fromUserID, toUserID); err != nil {
		return fmt.Errorf("dedupe reassigned participant: %w", err)
	}
	if _, err := exec.ExecContext(ctx,
		`UPDATE schedule_layer_participants p SET user_id = $4
		 FROM schedule_layers l
		 WHERE p.layer_id = l.id AND l.schedule_id = $1 AND p.workspace_id = $2 AND p.user_id = $3`,
		scheduleID, workspaceID, fromUserID, toUserID); err != nil {
		return fmt.Errorf("reassign participant: %w", err)
	}
	if _, err := exec.ExecContext(ctx,
		`UPDATE schedule_overrides SET user_id = $4
		 WHERE schedule_id = $1 AND workspace_id = $2 AND user_id = $3 AND ends_at > now()`,
		scheduleID, workspaceID, fromUserID, toUserID); err != nil {
		return fmt.Errorf("reassign override: %w", err)
	}
	return nil
}
