package escalation_run

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

const nextRoundRobinSQL = `
INSERT INTO escalation_rr_state (policy_id, node_id, position)
VALUES ($1, $2, 0)
ON CONFLICT (policy_id, node_id) DO UPDATE SET position = escalation_rr_state.position + 1
RETURNING position`

const recentByPolicySQL = `
SELECT e.alert_id, a.title AS alert_title, e.started_at, e.ended_at, e.state, e.step_index,
       COALESCE(a.acked_by_label, '') AS by_label
FROM alert_escalations e
JOIN alerts a ON a.id = e.alert_id
WHERE e.policy_id = $1
ORDER BY e.started_at DESC
LIMIT $2`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.EscalationRun {
	return &repo{db: db}
}

type planJSON struct {
	Repeat            int               `json:"repeat"`
	AckTimeoutSeconds int               `json:"ackTimeoutSeconds"`
	Snapshot          json.RawMessage   `json:"snapshot"`
	Path              []levelJSON       `json:"path"`
	LaneChoices       map[string]string `json:"laneChoices,omitempty"`
	PolicySlug        string            `json:"policySlug"`
}

type levelJSON struct {
	ID          string       `json:"id"`
	Targets     []targetJSON `json:"targets"`
	Mode        string       `json:"mode"`
	WaitSeconds int          `json:"waitSeconds"`
}

type targetJSON struct {
	Type string `json:"type"`
	Ref  string `json:"ref"`
}

type snapNodeJSON struct {
	Type        string         `json:"type"`
	ID          string         `json:"id"`
	Targets     []targetJSON   `json:"targets,omitempty"`
	Mode        string         `json:"mode,omitempty"`
	WaitSeconds int            `json:"waitSeconds,omitempty"`
	On          string         `json:"on,omitempty"`
	Hours       *snapHoursJSON `json:"hours,omitempty"`
	Lanes       []snapLaneJSON `json:"lanes,omitempty"`
}

type snapLaneJSON struct {
	ID    string         `json:"id"`
	Key   string         `json:"key"`
	Nodes []snapNodeJSON `json:"nodes"`
}

type snapHoursJSON struct {
	Days        []int  `json:"days"`
	StartMinute int    `json:"startMinute"`
	EndMinute   int    `json:"endMinute"`
	Timezone    string `json:"timezone"`
}

func levelsToJSON(levels []entity.EscalationLevel) []levelJSON {
	out := make([]levelJSON, 0, len(levels))
	for _, l := range levels {
		targets := make([]targetJSON, 0, len(l.Targets))
		for _, t := range l.Targets {
			targets = append(targets, targetJSON{Type: string(t.Type), Ref: t.Ref})
		}
		out = append(out, levelJSON{ID: l.ID, Targets: targets, Mode: string(l.Mode), WaitSeconds: int(l.Wait / time.Second)})
	}
	return out
}

func levelsFromJSON(levels []levelJSON) []entity.EscalationLevel {
	out := make([]entity.EscalationLevel, 0, len(levels))
	for _, l := range levels {
		targets := make([]entity.EscalationTarget, 0, len(l.Targets))
		for _, t := range l.Targets {
			targets = append(targets, entity.EscalationTarget{Type: entity.EscalationTargetType(t.Type), Ref: t.Ref})
		}
		out = append(out, entity.EscalationLevel{ID: l.ID, Targets: targets, Mode: entity.NotifyMode(l.Mode), Wait: time.Duration(l.WaitSeconds) * time.Second})
	}
	return out
}

func snapToJSON(nodes []entity.EscalationNode) []snapNodeJSON {
	out := make([]snapNodeJSON, 0, len(nodes))
	for _, node := range nodes {
		switch {
		case node.Level != nil:
			targets := make([]targetJSON, 0, len(node.Level.Targets))
			for _, t := range node.Level.Targets {
				targets = append(targets, targetJSON{Type: string(t.Type), Ref: t.Ref})
			}
			out = append(out, snapNodeJSON{
				Type: "level", ID: node.Level.ID, Targets: targets,
				Mode: string(node.Level.Mode), WaitSeconds: int(node.Level.Wait / time.Second),
			})
		case node.Branch != nil:
			lanes := make([]snapLaneJSON, 0, len(node.Branch.Lanes))
			for _, lane := range node.Branch.Lanes {
				lanes = append(lanes, snapLaneJSON{ID: lane.ID, Key: lane.Key, Nodes: snapToJSON(lane.Nodes)})
			}
			n := snapNodeJSON{Type: "branch", ID: node.Branch.ID, On: string(node.Branch.On), Lanes: lanes}
			if node.Branch.On == entity.EscalationBranchHours {
				n.Hours = &snapHoursJSON{
					Days:        node.Branch.Hours.Days,
					StartMinute: node.Branch.Hours.StartMinute,
					EndMinute:   node.Branch.Hours.EndMinute,
					Timezone:    node.Branch.Hours.Timezone,
				}
			}
			out = append(out, n)
		}
	}
	return out
}

func snapFromJSON(nodes []snapNodeJSON) []entity.EscalationNode {
	out := make([]entity.EscalationNode, 0, len(nodes))
	for _, n := range nodes {
		switch n.Type {
		case "level":
			targets := make([]entity.EscalationTarget, 0, len(n.Targets))
			for _, t := range n.Targets {
				targets = append(targets, entity.EscalationTarget{Type: entity.EscalationTargetType(t.Type), Ref: t.Ref})
			}
			out = append(out, entity.EscalationNode{Level: &entity.EscalationLevel{
				ID: n.ID, Targets: targets, Mode: entity.NotifyMode(n.Mode), Wait: time.Duration(n.WaitSeconds) * time.Second,
			}})
		case "branch":
			lanes := make([]entity.EscalationLane, 0, len(n.Lanes))
			for _, lane := range n.Lanes {
				lanes = append(lanes, entity.EscalationLane{ID: lane.ID, Key: lane.Key, Nodes: snapFromJSON(lane.Nodes)})
			}
			branch := &entity.EscalationBranch{ID: n.ID, On: entity.EscalationBranchKind(n.On), Lanes: lanes}
			if n.Hours != nil {
				branch.Hours = entity.HoursWindow{
					Days: n.Hours.Days, StartMinute: n.Hours.StartMinute,
					EndMinute: n.Hours.EndMinute, Timezone: n.Hours.Timezone,
				}
			}
			out = append(out, entity.EscalationNode{Branch: branch})
		}
	}
	return out
}

func marshalPlan(run entity.EscalationRun) (types.JSON, error) {
	snapshot, err := json.Marshal(snapToJSON(run.Snapshot.Nodes))
	if err != nil {
		return nil, fmt.Errorf("encode escalation snapshot: %w", err)
	}
	raw, err := json.Marshal(planJSON{
		Repeat:            run.Snapshot.Repeat,
		AckTimeoutSeconds: int(run.Snapshot.AckTimeout / time.Second),
		Snapshot:          snapshot,
		Path:              levelsToJSON(run.Path),
		LaneChoices:       run.LaneChoices,
		PolicySlug:        run.PolicySlug,
	})
	if err != nil {
		return nil, fmt.Errorf("encode escalation plan: %w", err)
	}
	return types.JSON(raw), nil
}

func unmarshalPlan(raw types.JSON, run *entity.EscalationRun) {
	var plan planJSON
	if err := json.Unmarshal(raw, &plan); err != nil {
		return
	}
	var snapshot []snapNodeJSON
	_ = json.Unmarshal(plan.Snapshot, &snapshot)
	run.Snapshot = entity.EscalationSnapshot{
		Repeat:     plan.Repeat,
		AckTimeout: time.Duration(plan.AckTimeoutSeconds) * time.Second,
		Nodes:      snapFromJSON(snapshot),
	}
	run.Path = levelsFromJSON(plan.Path)
	run.LaneChoices = plan.LaneChoices
	run.PolicySlug = plan.PolicySlug
}

func toEntity(m *dbpostgres.AlertEscalation) entity.EscalationRun {
	out := entity.EscalationRun{
		ID:           m.ID,
		WorkspaceID:  m.WorkspaceID,
		AlertID:      m.AlertID,
		PolicyID:     m.PolicyID,
		State:        entity.EscalationRunState(m.State),
		Cycle:        m.Cycle,
		StepIndex:    m.StepIndex,
		NextAt:       m.NextAt.Time,
		AckedAt:      m.AckedAt.Time,
		AckExpiresAt: m.AckExpiresAt.Time,
		StartedAt:    m.StartedAt,
		EndedAt:      m.EndedAt.Time,
		UpdatedAt:    m.UpdatedAt,
	}
	unmarshalPlan(m.Plan, &out)
	return out
}

func (r *repo) Create(ctx context.Context, run entity.EscalationRun) (entity.EscalationRun, bool, error) {
	plan, err := marshalPlan(run)
	if err != nil {
		return entity.EscalationRun{}, false, err
	}
	m := &dbpostgres.AlertEscalation{
		WorkspaceID: run.WorkspaceID,
		AlertID:     run.AlertID,
		PolicyID:    run.PolicyID,
		State:       string(run.State),
		Cycle:       run.Cycle,
		StepIndex:   run.StepIndex,
		Plan:        plan,
		StartedAt:   run.StartedAt,
	}
	if !run.NextAt.IsZero() {
		m.NextAt = null.TimeFrom(run.NextAt)
	}
	cols := boil.Whitelist("workspace_id", "alert_id", "policy_id", "state", "cycle", "step_index", "plan", "next_at", "started_at")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			existing, getErr := r.GetByAlertID(ctx, run.AlertID)
			return existing, false, getErr
		}
		return entity.EscalationRun{}, false, fmt.Errorf("create escalation run: %w", err)
	}
	return toEntity(m), true, nil
}

func (r *repo) GetByAlertID(ctx context.Context, alertID string) (entity.EscalationRun, error) {
	m, err := dbpostgres.AlertEscalations(qm.Where("alert_id = ?", alertID)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.EscalationRun{}, entity.ErrEscalationRunNotFound
		}
		return entity.EscalationRun{}, fmt.Errorf("get escalation run: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) ListDue(ctx context.Context, now time.Time, limit int) ([]entity.EscalationRun, error) {
	if limit <= 0 {
		limit = entity.EscalationSweepBatch
	}
	rows, err := dbpostgres.AlertEscalations(
		qm.Where("state IN ('running', 'acked') AND next_at IS NOT NULL AND next_at <= ?", now),
		qm.OrderBy("next_at, id"),
		qm.Limit(limit),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list due escalations: %w", err)
	}
	out := make([]entity.EscalationRun, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEntity(m))
	}
	return out, nil
}

func (r *repo) SaveProgress(ctx context.Context, run entity.EscalationRun) (bool, error) {
	plan, err := marshalPlan(run)
	if err != nil {
		return false, err
	}
	values := dbpostgres.M{
		"cycle":      run.Cycle,
		"step_index": run.StepIndex,
		"plan":       plan,
		"updated_at": time.Now().UTC(),
	}
	if run.NextAt.IsZero() {
		values["next_at"] = nil
	} else {
		values["next_at"] = run.NextAt
	}
	affected, err := dbpostgres.AlertEscalations(
		qm.Where("id = ? AND state = ?", run.ID, string(entity.EscalationRunning)),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("save escalation progress: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) MarkAcked(ctx context.Context, alertID string, ackedAt, expiresAt time.Time) (bool, error) {
	values := dbpostgres.M{
		"state":      string(entity.EscalationAcked),
		"acked_at":   ackedAt,
		"updated_at": ackedAt,
	}
	if expiresAt.IsZero() {
		values["ack_expires_at"] = nil
		values["next_at"] = nil
	} else {
		values["ack_expires_at"] = expiresAt
		values["next_at"] = expiresAt
	}
	affected, err := dbpostgres.AlertEscalations(
		qm.Where("alert_id = ? AND state = ?", alertID, string(entity.EscalationRunning)),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("mark escalation acked: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) MarkResolved(ctx context.Context, alertIDs []string, at time.Time) (int, error) {
	if len(alertIDs) == 0 {
		return 0, nil
	}
	values := dbpostgres.M{
		"state":      string(entity.EscalationResolved),
		"ended_at":   at,
		"next_at":    nil,
		"updated_at": at,
	}
	affected, err := dbpostgres.AlertEscalations(
		qm.WhereIn("state IN ?", string(entity.EscalationRunning), string(entity.EscalationAcked)),
		qm.WhereIn("alert_id IN ?", anySlice(alertIDs)...),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return 0, fmt.Errorf("mark escalations resolved: %w", err)
	}
	return int(affected), nil
}

func (r *repo) Resume(ctx context.Context, runID string, at time.Time) (bool, error) {
	values := dbpostgres.M{
		"state":          string(entity.EscalationRunning),
		"acked_at":       nil,
		"ack_expires_at": nil,
		"next_at":        at,
		"updated_at":     at,
	}
	affected, err := dbpostgres.AlertEscalations(
		qm.Where("id = ? AND state = ?", runID, string(entity.EscalationAcked)),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("resume escalation: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) Finish(ctx context.Context, runID string, state entity.EscalationRunState, at time.Time) (bool, error) {
	values := dbpostgres.M{
		"state":      string(state),
		"ended_at":   at,
		"next_at":    nil,
		"updated_at": at,
	}
	affected, err := dbpostgres.AlertEscalations(
		qm.Where("id = ?", runID),
		qm.WhereIn("state IN ?", string(entity.EscalationRunning), string(entity.EscalationAcked)),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return false, fmt.Errorf("finish escalation: %w", err)
	}
	return affected > 0, nil
}

func (r *repo) NextRoundRobin(ctx context.Context, policyID, nodeID string) (int, error) {
	var position int
	if err := r.db.Querier(ctx).QueryRowContext(ctx, nextRoundRobinSQL, policyID, nodeID).Scan(&position); err != nil {
		return 0, fmt.Errorf("advance round robin: %w", err)
	}
	return position, nil
}

func (r *repo) RecentByPolicy(ctx context.Context, policyID string, limit int) ([]entity.EscalationRecent, error) {
	if limit <= 0 {
		limit = entity.EscalationRecentLimit
	}
	var rows []struct {
		AlertID    string    `boil:"alert_id"`
		AlertTitle string    `boil:"alert_title"`
		StartedAt  time.Time `boil:"started_at"`
		EndedAt    null.Time `boil:"ended_at"`
		State      string    `boil:"state"`
		StepIndex  int       `boil:"step_index"`
		ByLabel    string    `boil:"by_label"`
	}
	if err := queries.Raw(recentByPolicySQL, policyID, limit).Bind(ctx, r.db.Querier(ctx), &rows); err != nil {
		return nil, fmt.Errorf("list recent escalations: %w", err)
	}
	out := make([]entity.EscalationRecent, 0, len(rows))
	for _, row := range rows {
		run := entity.EscalationRun{State: entity.EscalationRunState(row.State), StepIndex: row.StepIndex}
		out = append(out, entity.EscalationRecent{
			AlertID:    row.AlertID,
			AlertTitle: row.AlertTitle,
			StartedAt:  row.StartedAt,
			EndedAt:    row.EndedAt.Time,
			State:      run.State,
			Outcome:    run.Outcome(),
			ByLabel:    row.ByLabel,
			StepIndex:  row.StepIndex,
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
