package notification_run

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
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

func New(db postgres.Client) repository.NotificationRun {
	return &repo{db: db}
}

type planJSON struct {
	Urgency    string     `json:"urgency"`
	QuietHours quietJSON  `json:"quietHours"`
	Steps      []stepJSON `json:"steps"`
}

type quietJSON struct {
	Enabled     bool   `json:"enabled"`
	Days        []int  `json:"days"`
	StartMinute int    `json:"startMinute"`
	EndMinute   int    `json:"endMinute"`
	Timezone    string `json:"timezone"`
}

type stepJSON struct {
	Channel      string `json:"channel"`
	DelaySeconds int    `json:"delaySeconds"`
	ChannelID    string `json:"channelId,omitempty"`
	Detail       string `json:"detail,omitempty"`
}

func marshalPlan(plan entity.NotificationPlan) (types.JSON, error) {
	steps := make([]stepJSON, 0, len(plan.Steps))
	for _, s := range plan.Steps {
		steps = append(steps, stepJSON{
			Channel:      string(s.Channel),
			DelaySeconds: int(s.Delay / time.Second),
			ChannelID:    s.ChannelID,
			Detail:       s.Detail,
		})
	}
	raw, err := json.Marshal(planJSON{
		Urgency: string(plan.Urgency),
		QuietHours: quietJSON{
			Enabled:     plan.QuietHours.Enabled,
			Days:        plan.QuietHours.Window.Days,
			StartMinute: plan.QuietHours.Window.StartMinute,
			EndMinute:   plan.QuietHours.Window.EndMinute,
			Timezone:    plan.QuietHours.Window.Timezone,
		},
		Steps: steps,
	})
	if err != nil {
		return nil, fmt.Errorf("encode notification plan: %w", err)
	}
	return types.JSON(raw), nil
}

func unmarshalPlan(raw types.JSON, plan *entity.NotificationPlan) error {
	if len(raw) == 0 {
		return nil
	}
	var pj planJSON
	if err := json.Unmarshal(raw, &pj); err != nil {
		return fmt.Errorf("decode notification plan: %w", err)
	}
	plan.Urgency = entity.NotifyUrgency(pj.Urgency)
	plan.QuietHours = entity.QuietHours{
		Enabled: pj.QuietHours.Enabled,
		Window: entity.HoursWindow{
			Days:        pj.QuietHours.Days,
			StartMinute: pj.QuietHours.StartMinute,
			EndMinute:   pj.QuietHours.EndMinute,
			Timezone:    pj.QuietHours.Timezone,
		},
	}
	plan.Steps = make([]entity.NotificationPlanStep, 0, len(pj.Steps))
	for _, s := range pj.Steps {
		plan.Steps = append(plan.Steps, entity.NotificationPlanStep{
			Channel:   entity.ChannelType(s.Channel),
			Delay:     time.Duration(s.DelaySeconds) * time.Second,
			ChannelID: s.ChannelID,
			Detail:    s.Detail,
		})
	}
	return nil
}

func toEntity(m *dbpostgres.NotificationRun) (entity.NotificationRun, error) {
	out := entity.NotificationRun{
		ID:              m.ID,
		WorkspaceID:     m.WorkspaceID,
		AlertID:         m.AlertID,
		UserID:          m.UserID,
		EscalationID:    m.EscalationID.String,
		EscalationCycle: m.EscalationCycle,
		Level:           m.Level,
		PolicySlug:      m.PolicySlug,
		Label:           m.Label,
		Urgency:         entity.NotifyUrgency(m.Urgency),
		State:           entity.NotifyRunState(m.State),
		StopReason:      entity.NotifyStopReason(m.StopReason),
		StepIndex:       m.StepIndex,
		NextAt:          m.NextAt.Time,
		StartedAt:       m.StartedAt,
		EndedAt:         m.EndedAt.Time,
		UpdatedAt:       m.UpdatedAt,
	}
	if err := unmarshalPlan(m.Plan, &out.Plan); err != nil {
		return entity.NotificationRun{}, err
	}
	return out, nil
}

const supersedeSQL = `
UPDATE notification_runs
   SET state = 'stopped', stop_reason = 'superseded', next_at = NULL, ended_at = $1, updated_at = $1
 WHERE alert_id = $2 AND user_id = $3 AND id <> $4 AND state = 'running'`

func (r *repo) Create(ctx context.Context, run entity.NotificationRun) (entity.NotificationRun, bool, error) {
	plan, err := marshalPlan(run.Plan)
	if err != nil {
		return entity.NotificationRun{}, false, err
	}
	m := &dbpostgres.NotificationRun{
		WorkspaceID:     run.WorkspaceID,
		AlertID:         run.AlertID,
		UserID:          run.UserID,
		EscalationCycle: run.EscalationCycle,
		Level:           run.Level,
		PolicySlug:      run.PolicySlug,
		Label:           run.Label,
		Urgency:         string(run.Urgency),
		State:           string(run.State),
		StepIndex:       run.StepIndex,
		Plan:            plan,
		StartedAt:       run.StartedAt,
	}
	if run.EscalationID != "" {
		m.EscalationID = null.StringFrom(run.EscalationID)
	}
	if !run.NextAt.IsZero() {
		m.NextAt = null.TimeFrom(run.NextAt)
	}
	cols := boil.Whitelist("workspace_id", "alert_id", "user_id", "escalation_id", "escalation_cycle",
		"level", "policy_slug", "label", "urgency", "state", "step_index", "plan", "next_at", "started_at")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			existing, getErr := r.getByPage(ctx, run)
			return existing, false, getErr
		}
		return entity.NotificationRun{}, false, fmt.Errorf("create notification run: %w", err)
	}
	if _, err := r.db.Querier(ctx).ExecContext(ctx, supersedeSQL, run.StartedAt, run.AlertID, run.UserID, m.ID); err != nil {
		return entity.NotificationRun{}, false, fmt.Errorf("supersede notification runs: %w", err)
	}
	out, err := toEntity(m)
	return out, true, err
}

func (r *repo) getByPage(ctx context.Context, run entity.NotificationRun) (entity.NotificationRun, error) {
	m, err := dbpostgres.NotificationRuns(
		qm.Where("alert_id = ? AND user_id = ? AND level = ? AND escalation_cycle = ?",
			run.AlertID, run.UserID, run.Level, run.EscalationCycle),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.NotificationRun{}, entity.ErrNotificationRunNotFound
		}
		return entity.NotificationRun{}, fmt.Errorf("get notification run: %w", err)
	}
	return toEntity(m)
}

func (r *repo) GetByID(ctx context.Context, id string) (entity.NotificationRun, error) {
	m, err := dbpostgres.NotificationRuns(qm.Where("id = ?", id)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.NotificationRun{}, entity.ErrNotificationRunNotFound
		}
		return entity.NotificationRun{}, fmt.Errorf("get notification run: %w", err)
	}
	return toEntity(m)
}

func (r *repo) ListDue(ctx context.Context, now time.Time, limit int) ([]entity.NotificationRun, error) {
	if limit <= 0 {
		limit = entity.NotificationRunSweepBatch
	}
	rows, err := dbpostgres.NotificationRuns(
		qm.Where("state = 'running' AND next_at IS NOT NULL AND next_at <= ?", now),
		qm.OrderBy("next_at, id"),
		qm.Limit(limit),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list due notification runs: %w", err)
	}
	out := make([]entity.NotificationRun, 0, len(rows))
	for _, m := range rows {
		run, err := toEntity(m)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, nil
}

func (r *repo) SaveProgress(ctx context.Context, run entity.NotificationRun) (bool, error) {
	values := dbpostgres.M{
		"step_index":  run.StepIndex,
		"state":       string(run.State),
		"stop_reason": string(run.StopReason),
		"updated_at":  time.Now().UTC(),
	}
	if run.NextAt.IsZero() {
		values["next_at"] = nil
	} else {
		values["next_at"] = run.NextAt
	}
	if run.EndedAt.IsZero() {
		values["ended_at"] = nil
	} else {
		values["ended_at"] = run.EndedAt
	}
	affected, err := dbpostgres.NotificationRuns(
		qm.Where("id = ? AND state = ?", run.ID, string(entity.NotifyRunRunning)),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("save notification progress: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) Reschedule(ctx context.Context, runID string, stepIndex int, at time.Time) (bool, error) {
	values := dbpostgres.M{"next_at": at, "updated_at": time.Now().UTC()}
	affected, err := dbpostgres.NotificationRuns(
		qm.Where("id = ? AND state = 'running' AND step_index = ?", runID, stepIndex),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("reschedule notification run: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) StopByAlerts(ctx context.Context, workspaceID string, alertIDs []string, reason entity.NotifyStopReason, at time.Time) (int, error) {
	if len(alertIDs) == 0 {
		return 0, nil
	}
	values := dbpostgres.M{
		"state":       string(entity.NotifyRunStopped),
		"stop_reason": string(reason),
		"next_at":     nil,
		"ended_at":    at,
		"updated_at":  at,
	}
	affected, err := dbpostgres.NotificationRuns(
		qm.Where("workspace_id = ? AND state = 'running'", workspaceID),
		qm.WhereIn("alert_id IN ?", anySlice(alertIDs)...),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return 0, fmt.Errorf("stop notification runs: %w", err)
	}
	return int(affected), nil
}

func (r *repo) ListByAlert(ctx context.Context, alertID string) ([]entity.NotificationRun, error) {
	rows, err := dbpostgres.NotificationRuns(
		qm.Where("alert_id = ?", alertID),
		qm.OrderBy("started_at DESC, id"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list notification runs: %w", err)
	}
	out := make([]entity.NotificationRun, 0, len(rows))
	for _, m := range rows {
		run, err := toEntity(m)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, nil
}

func (r *repo) AppendAttempt(ctx context.Context, attempt entity.NotificationAttempt) error {
	m := &dbpostgres.NotificationAttempt{
		RunID:             attempt.RunID,
		WorkspaceID:       attempt.WorkspaceID,
		AlertID:           attempt.AlertID,
		UserID:            attempt.UserID,
		StepIndex:         attempt.StepIndex,
		ChannelType:       string(attempt.Channel),
		Detail:            attempt.Detail,
		Outcome:           string(attempt.Outcome),
		ProviderMessageID: attempt.ProviderMessageID,
		ErrorDetail:       attempt.ErrorDetail,
	}
	if attempt.ChannelID != "" {
		m.ChannelID = null.StringFrom(attempt.ChannelID)
	}
	if !attempt.At.IsZero() {
		m.At = attempt.At
	}
	cols := boil.Whitelist("run_id", "workspace_id", "alert_id", "user_id", "step_index",
		"channel_type", "channel_id", "detail", "outcome", "provider_message_id", "error_detail", "at")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		return fmt.Errorf("append notification attempt: %w", err)
	}
	return nil
}

func (r *repo) ListAttempts(ctx context.Context, alertID string, limit int) ([]entity.NotificationAttempt, error) {
	if limit <= 0 {
		limit = entity.AlertTimelineLimit
	}
	rows, err := dbpostgres.NotificationAttempts(
		qm.Where("alert_id = ?", alertID),
		qm.OrderBy("at, id"),
		qm.Limit(limit),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list notification attempts: %w", err)
	}
	out := make([]entity.NotificationAttempt, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.NotificationAttempt{
			ID:                m.ID,
			RunID:             m.RunID,
			WorkspaceID:       m.WorkspaceID,
			AlertID:           m.AlertID,
			UserID:            m.UserID,
			StepIndex:         m.StepIndex,
			Channel:           entity.ChannelType(m.ChannelType),
			ChannelID:         m.ChannelID.String,
			Detail:            m.Detail,
			Outcome:           entity.NotifyOutcome(m.Outcome),
			ProviderMessageID: m.ProviderMessageID,
			ErrorDetail:       m.ErrorDetail,
			At:                m.At,
		})
	}
	return out, nil
}

func anySlice(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}
