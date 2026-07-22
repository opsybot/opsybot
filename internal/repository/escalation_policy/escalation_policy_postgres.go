package escalation_policy

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
	"github.com/opsybot/opsybot/internal/pkg/secretbox"
	"github.com/opsybot/opsybot/internal/repository"
)

const routedCountsSQL = `
SELECT escalation_policy_id AS policy_id, count(*)::int AS routed
FROM alerts
WHERE workspace_id = $1 AND escalation_policy_id IS NOT NULL AND parent_alert_id IS NULL
GROUP BY 1`

type repo struct {
	db  postgres.Client
	box secretbox.Client
}

func New(db postgres.Client, box secretbox.Client) repository.EscalationPolicy {
	return &repo{db: db, box: box}
}

type targetJSON struct {
	Type string `json:"type"`
	Ref  string `json:"ref"`
}

type hoursJSON struct {
	Days        []int  `json:"days"`
	StartMinute int    `json:"startMinute"`
	EndMinute   int    `json:"endMinute"`
	Timezone    string `json:"timezone"`
}

type laneJSON struct {
	ID    string     `json:"id"`
	Key   string     `json:"key"`
	Nodes []nodeJSON `json:"nodes"`
}

type nodeJSON struct {
	Type        string       `json:"type"`
	ID          string       `json:"id"`
	Targets     []targetJSON `json:"targets,omitempty"`
	Mode        string       `json:"mode,omitempty"`
	WaitSeconds int          `json:"waitSeconds,omitempty"`
	On          string       `json:"on,omitempty"`
	Hours       *hoursJSON   `json:"hours,omitempty"`
	Lanes       []laneJSON   `json:"lanes,omitempty"`
}

func nodesToJSON(nodes []entity.EscalationNode) []nodeJSON {
	out := make([]nodeJSON, 0, len(nodes))
	for _, node := range nodes {
		switch {
		case node.Level != nil:
			targets := make([]targetJSON, 0, len(node.Level.Targets))
			for _, t := range node.Level.Targets {
				targets = append(targets, targetJSON{Type: string(t.Type), Ref: t.Ref})
			}
			out = append(out, nodeJSON{
				Type:        "level",
				ID:          node.Level.ID,
				Targets:     targets,
				Mode:        string(node.Level.Mode),
				WaitSeconds: int(node.Level.Wait / time.Second),
			})
		case node.Branch != nil:
			lanes := make([]laneJSON, 0, len(node.Branch.Lanes))
			for _, lane := range node.Branch.Lanes {
				lanes = append(lanes, laneJSON{ID: lane.ID, Key: lane.Key, Nodes: nodesToJSON(lane.Nodes)})
			}
			n := nodeJSON{Type: "branch", ID: node.Branch.ID, On: string(node.Branch.On), Lanes: lanes}
			if node.Branch.On == entity.EscalationBranchHours {
				n.Hours = &hoursJSON{
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

func nodesFromJSON(nodes []nodeJSON) []entity.EscalationNode {
	out := make([]entity.EscalationNode, 0, len(nodes))
	for _, n := range nodes {
		switch n.Type {
		case "level":
			targets := make([]entity.EscalationTarget, 0, len(n.Targets))
			for _, t := range n.Targets {
				targets = append(targets, entity.EscalationTarget{Type: entity.EscalationTargetType(t.Type), Ref: t.Ref})
			}
			out = append(out, entity.EscalationNode{Level: &entity.EscalationLevel{
				ID:      n.ID,
				Targets: targets,
				Mode:    entity.NotifyMode(n.Mode),
				Wait:    time.Duration(n.WaitSeconds) * time.Second,
			}})
		case "branch":
			lanes := make([]entity.EscalationLane, 0, len(n.Lanes))
			for _, lane := range n.Lanes {
				lanes = append(lanes, entity.EscalationLane{ID: lane.ID, Key: lane.Key, Nodes: nodesFromJSON(lane.Nodes)})
			}
			branch := &entity.EscalationBranch{ID: n.ID, On: entity.EscalationBranchKind(n.On), Lanes: lanes}
			if n.Hours != nil {
				branch.Hours = entity.HoursWindow{
					Days:        n.Hours.Days,
					StartMinute: n.Hours.StartMinute,
					EndMinute:   n.Hours.EndMinute,
					Timezone:    n.Hours.Timezone,
				}
			}
			out = append(out, entity.EscalationNode{Branch: branch})
		}
	}
	return out
}

func marshalDefinition(nodes []entity.EscalationNode) (types.JSON, error) {
	raw, err := json.Marshal(nodesToJSON(nodes))
	if err != nil {
		return nil, fmt.Errorf("encode policy definition: %w", err)
	}
	return types.JSON(raw), nil
}

func unmarshalDefinition(raw types.JSON) []entity.EscalationNode {
	var nodes []nodeJSON
	if err := json.Unmarshal(raw, &nodes); err != nil {
		return nil
	}
	return nodesFromJSON(nodes)
}

func (r *repo) toEntity(m *dbpostgres.EscalationPolicy) entity.EscalationPolicy {
	out := entity.EscalationPolicy{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		Slug:        m.Slug,
		Name:        m.Name,
		TeamID:      m.TeamID.String,
		Repeat:      m.RepeatCount,
		AckTimeout:  time.Duration(m.AckTimeoutSeconds) * time.Second,
		Nodes:       unmarshalDefinition(m.Definition),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.R != nil {
		if team := m.R.GetTeam(); team != nil {
			out.TeamSlug = team.Slug
		}
	}
	return out
}

func (r *repo) List(ctx context.Context, workspaceID string) ([]entity.EscalationPolicy, error) {
	rows, err := dbpostgres.EscalationPolicies(
		qm.Where("workspace_id = ?", workspaceID),
		qm.Load(dbpostgres.EscalationPolicyRels.Team),
		qm.OrderBy("name, id"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list escalation policies: %w", err)
	}
	out := make([]entity.EscalationPolicy, 0, len(rows))
	for _, m := range rows {
		out = append(out, r.toEntity(m))
	}
	return out, nil
}

func (r *repo) find(ctx context.Context, where qm.QueryMod) (entity.EscalationPolicy, error) {
	m, err := dbpostgres.EscalationPolicies(where, qm.Load(dbpostgres.EscalationPolicyRels.Team)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.EscalationPolicy{}, entity.ErrEscalationPolicyNotFound
		}
		return entity.EscalationPolicy{}, fmt.Errorf("get escalation policy: %w", err)
	}
	return r.toEntity(m), nil
}

func (r *repo) GetBySlug(ctx context.Context, workspaceID, slug string) (entity.EscalationPolicy, error) {
	return r.find(ctx, qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug))
}

func (r *repo) GetByID(ctx context.Context, workspaceID, id string) (entity.EscalationPolicy, error) {
	return r.find(ctx, qm.Where("workspace_id = ? AND id = ?", workspaceID, id))
}

func (r *repo) Create(ctx context.Context, p entity.EscalationPolicy) (entity.EscalationPolicy, error) {
	definition, err := marshalDefinition(p.Nodes)
	if err != nil {
		return entity.EscalationPolicy{}, err
	}
	m := &dbpostgres.EscalationPolicy{
		WorkspaceID:       p.WorkspaceID,
		Slug:              p.Slug,
		Name:              p.Name,
		RepeatCount:       p.Repeat,
		AckTimeoutSeconds: int(p.AckTimeout / time.Second),
		Definition:        definition,
	}
	if p.TeamID != "" {
		m.TeamID = null.StringFrom(p.TeamID)
	}
	cols := boil.Whitelist("workspace_id", "slug", "name", "team_id", "repeat_count", "ack_timeout_seconds", "definition")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.EscalationPolicy{}, entity.ErrEscalationPolicySlugTaken
		}
		return entity.EscalationPolicy{}, fmt.Errorf("create escalation policy: %w", err)
	}
	if err := r.rebuildTargets(ctx, m.ID, p.Nodes); err != nil {
		return entity.EscalationPolicy{}, err
	}
	return r.GetByID(ctx, p.WorkspaceID, m.ID)
}

func (r *repo) Update(ctx context.Context, p entity.EscalationPolicy) (entity.EscalationPolicy, error) {
	definition, err := marshalDefinition(p.Nodes)
	if err != nil {
		return entity.EscalationPolicy{}, err
	}
	values := dbpostgres.M{
		"name":                p.Name,
		"repeat_count":        p.Repeat,
		"ack_timeout_seconds": int(p.AckTimeout / time.Second),
		"definition":          definition,
		"updated_at":          time.Now().UTC(),
	}
	if p.TeamID != "" {
		values["team_id"] = p.TeamID
	} else {
		values["team_id"] = nil
	}
	affected, err := dbpostgres.EscalationPolicies(
		qm.Where("workspace_id = ? AND id = ?", p.WorkspaceID, p.ID),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return entity.EscalationPolicy{}, fmt.Errorf("update escalation policy: %w", err)
	}
	if affected == 0 {
		return entity.EscalationPolicy{}, entity.ErrEscalationPolicyNotFound
	}
	if err := r.rebuildTargets(ctx, p.ID, p.Nodes); err != nil {
		return entity.EscalationPolicy{}, err
	}
	return r.GetByID(ctx, p.WorkspaceID, p.ID)
}

func (r *repo) rebuildTargets(ctx context.Context, policyID string, nodes []entity.EscalationNode) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.EscalationPolicyTargets(qm.Where("policy_id = ?", policyID)).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("clear policy targets: %w", err)
	}
	for nodeID, targets := range targetsByNode(nodes) {
		for _, t := range targets {
			row := &dbpostgres.EscalationPolicyTarget{
				PolicyID:   policyID,
				NodeID:     nodeID,
				TargetType: string(t.Type),
				TargetRef:  t.Ref,
			}
			if err := row.Insert(ctx, exec, boil.Whitelist("policy_id", "node_id", "target_type", "target_ref")); err != nil {
				return fmt.Errorf("insert policy target: %w", err)
			}
		}
	}
	return nil
}

func targetsByNode(nodes []entity.EscalationNode) map[string][]entity.EscalationTarget {
	out := map[string][]entity.EscalationTarget{}
	var walk func(nodes []entity.EscalationNode)
	walk = func(nodes []entity.EscalationNode) {
		for _, node := range nodes {
			switch {
			case node.Level != nil:
				out[node.Level.ID] = append(out[node.Level.ID], node.Level.Targets...)
			case node.Branch != nil:
				for _, lane := range node.Branch.Lanes {
					walk(lane.Nodes)
				}
			}
		}
	}
	walk(nodes)
	return out
}

func (r *repo) Delete(ctx context.Context, workspaceID, id string) error {
	affected, err := dbpostgres.EscalationPolicies(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
	).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("delete escalation policy: %w", err)
	}
	if affected == 0 {
		return entity.ErrEscalationPolicyNotFound
	}
	return nil
}

func (r *repo) Refs(ctx context.Context, workspaceID, policyID string) (entity.EscalationPolicyRefs, error) {
	exec := r.db.Querier(ctx)
	routes, err := dbpostgres.AlertRoutes(qm.Where("workspace_id = ? AND escalation_policy_id = ?", workspaceID, policyID)).Count(ctx, exec)
	if err != nil {
		return entity.EscalationPolicyRefs{}, fmt.Errorf("count policy routes: %w", err)
	}
	monitors, err := dbpostgres.AlertMonitors(qm.Where("workspace_id = ? AND escalation_policy_id = ?", workspaceID, policyID)).Count(ctx, exec)
	if err != nil {
		return entity.EscalationPolicyRefs{}, fmt.Errorf("count policy monitors: %w", err)
	}
	isDefault, err := dbpostgres.AlertSettings(qm.Where("workspace_id = ? AND default_escalation_policy_id = ?", workspaceID, policyID)).Exists(ctx, exec)
	if err != nil {
		return entity.EscalationPolicyRefs{}, fmt.Errorf("check policy default: %w", err)
	}
	active, err := dbpostgres.AlertEscalations(
		qm.Where("workspace_id = ? AND policy_id = ?", workspaceID, policyID),
		qm.WhereIn("state IN ?", string(entity.EscalationRunning), string(entity.EscalationAcked)),
	).Count(ctx, exec)
	if err != nil {
		return entity.EscalationPolicyRefs{}, fmt.Errorf("count policy active runs: %w", err)
	}
	return entity.EscalationPolicyRefs{Routes: int(routes), Monitors: int(monitors), Default: isDefault, ActiveRuns: int(active)}, nil
}

func (r *repo) RoutedCounts(ctx context.Context, workspaceID string) (map[string]int, error) {
	var rows []struct {
		PolicyID string `boil:"policy_id"`
		Routed   int    `boil:"routed"`
	}
	err := queries.Raw(routedCountsSQL, workspaceID).Bind(ctx, r.db.Querier(ctx), &rows)
	if err != nil {
		return nil, fmt.Errorf("count routed alerts: %w", err)
	}
	out := make(map[string]int, len(rows))
	for _, row := range rows {
		out[row.PolicyID] = row.Routed
	}
	return out, nil
}

func (r *repo) listReferencing(ctx context.Context, workspaceID, targetType, ref string) ([]entity.EscalationPolicy, error) {
	rows, err := dbpostgres.EscalationPolicies(
		qm.InnerJoin("escalation_policy_targets t ON t.policy_id = escalation_policies.id"),
		qm.Where("escalation_policies.workspace_id = ? AND t.target_type = ? AND t.target_ref = ?", workspaceID, targetType, ref),
		qm.GroupBy("escalation_policies.id"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list referencing policies: %w", err)
	}
	out := make([]entity.EscalationPolicy, 0, len(rows))
	for _, m := range rows {
		out = append(out, r.toEntity(m))
	}
	return out, nil
}

func (r *repo) ListReferencingUser(ctx context.Context, workspaceID, userID string) ([]entity.EscalationPolicy, error) {
	return r.listReferencing(ctx, workspaceID, string(entity.EscalationTargetPerson), userID)
}

func (r *repo) ListReferencingSchedule(ctx context.Context, workspaceID, scheduleID string) ([]entity.EscalationPolicy, error) {
	return r.listReferencing(ctx, workspaceID, string(entity.EscalationTargetSchedule), scheduleID)
}

func (r *repo) ListReferencingTeam(ctx context.Context, workspaceID, teamID string) ([]entity.EscalationPolicy, error) {
	return r.listReferencing(ctx, workspaceID, string(entity.EscalationTargetTeam), teamID)
}

func (r *repo) ListReferencingWebhook(ctx context.Context, workspaceID, webhookID string) ([]entity.EscalationPolicy, error) {
	return r.listReferencing(ctx, workspaceID, string(entity.EscalationTargetWebhook), webhookID)
}

func (r *repo) webhookToEntity(m *dbpostgres.EscalationWebhook) entity.EscalationWebhook {
	out := entity.EscalationWebhook{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		Slug:        m.Slug,
		Name:        m.Name,
		URL:         m.URL,
		CreatedAt:   m.CreatedAt,
	}
	if len(m.Secret) > 0 && r.box.Enabled() {
		if plain, err := r.box.Decrypt(m.Secret); err == nil {
			out.Secret = string(plain)
		}
	}
	return out
}

func (r *repo) ListWebhooks(ctx context.Context, workspaceID string) ([]entity.EscalationWebhook, error) {
	rows, err := dbpostgres.EscalationWebhooks(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("name, id"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list escalation webhooks: %w", err)
	}
	out := make([]entity.EscalationWebhook, 0, len(rows))
	for _, m := range rows {
		out = append(out, r.webhookToEntity(m))
	}
	return out, nil
}

func (r *repo) GetWebhook(ctx context.Context, workspaceID, slug string) (entity.EscalationWebhook, error) {
	m, err := dbpostgres.EscalationWebhooks(
		qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.EscalationWebhook{}, entity.ErrEscalationWebhookNotFound
		}
		return entity.EscalationWebhook{}, fmt.Errorf("get escalation webhook: %w", err)
	}
	return r.webhookToEntity(m), nil
}

func (r *repo) CreateWebhook(ctx context.Context, workspaceID string, in entity.NewEscalationWebhook, secret string) (entity.EscalationWebhook, error) {
	m := &dbpostgres.EscalationWebhook{
		WorkspaceID: workspaceID,
		Slug:        in.Slug,
		Name:        in.Name,
		URL:         in.URL,
		Secret:      []byte{},
	}
	if secret != "" {
		if !r.box.Enabled() {
			return entity.EscalationWebhook{}, entity.ErrEscalationSecretUnavailable
		}
		sealed, err := r.box.Encrypt([]byte(secret))
		if err != nil {
			return entity.EscalationWebhook{}, fmt.Errorf("seal webhook secret: %w", err)
		}
		m.Secret = sealed
	}
	cols := boil.Whitelist("workspace_id", "slug", "name", "url", "secret")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.EscalationWebhook{}, entity.ErrEscalationWebhookSlugTaken
		}
		return entity.EscalationWebhook{}, fmt.Errorf("create escalation webhook: %w", err)
	}
	out := r.webhookToEntity(m)
	out.Secret = secret
	return out, nil
}

func (r *repo) UpdateWebhook(ctx context.Context, workspaceID, slug string, in entity.NewEscalationWebhook) (entity.EscalationWebhook, error) {
	values := dbpostgres.M{"name": in.Name, "url": in.URL, "updated_at": time.Now().UTC()}
	affected, err := dbpostgres.EscalationWebhooks(
		qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return entity.EscalationWebhook{}, fmt.Errorf("update escalation webhook: %w", err)
	}
	if affected == 0 {
		return entity.EscalationWebhook{}, entity.ErrEscalationWebhookNotFound
	}
	return r.GetWebhook(ctx, workspaceID, slug)
}

func (r *repo) DeleteWebhook(ctx context.Context, workspaceID, slug string) error {
	affected, err := dbpostgres.EscalationWebhooks(
		qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug),
	).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("delete escalation webhook: %w", err)
	}
	if affected == 0 {
		return entity.ErrEscalationWebhookNotFound
	}
	return nil
}
