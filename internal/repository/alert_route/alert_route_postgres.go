package alert_route

import (
	"context"
	"database/sql"
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

func New(db postgres.Client) repository.AlertRoute {
	return &repo{db: db}
}

func (r *repo) List(ctx context.Context, workspaceID string) ([]entity.AlertRoute, error) {
	exec := r.db.Querier(ctx)
	rows, err := dbpostgres.AlertRoutes(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("position, id"),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list alert routes: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}

	ids := make([]any, 0, len(rows))
	for _, m := range rows {
		ids = append(ids, m.ID)
	}
	conds, err := dbpostgres.AlertRouteConditions(
		qm.WhereIn("route_id IN ?", ids...),
		qm.OrderBy("position, id"),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list route conditions: %w", err)
	}
	byRoute := make(map[string][]entity.RouteCondition, len(rows))
	for _, c := range conds {
		byRoute[c.RouteID] = append(byRoute[c.RouteID], entity.RouteCondition{
			Field: c.Field,
			Op:    entity.ConditionOp(c.Op),
			Value: c.Value,
		})
	}

	slugs, err := r.policySlugs(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	out := make([]entity.AlertRoute, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.AlertRoute{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			Position:    m.Position,
			PolicyID:    m.EscalationPolicyID,
			PolicySlug:  slugs[m.EscalationPolicyID],
			Conditions:  byRoute[m.ID],
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		})
	}
	return out, nil
}

func (r *repo) policySlugs(ctx context.Context, workspaceID string) (map[string]string, error) {
	rows, err := dbpostgres.EscalationPolicies(qm.Where("workspace_id = ?", workspaceID)).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list policy slugs: %w", err)
	}
	out := make(map[string]string, len(rows))
	for _, m := range rows {
		out[m.ID] = m.Slug
	}
	return out, nil
}

func (r *repo) Create(ctx context.Context, workspaceID string, in entity.NewAlertRoute) (entity.AlertRoute, error) {
	exec := r.db.Querier(ctx)
	next, err := dbpostgres.AlertRoutes(qm.Where("workspace_id = ?", workspaceID)).Count(ctx, exec)
	if err != nil {
		return entity.AlertRoute{}, fmt.Errorf("count alert routes: %w", err)
	}

	m := &dbpostgres.AlertRoute{WorkspaceID: workspaceID, EscalationPolicyID: in.PolicyID, Position: int(next)}
	if err := m.Insert(ctx, exec, boil.Whitelist("workspace_id", "escalation_policy_id", "position")); err != nil {
		return entity.AlertRoute{}, fmt.Errorf("create alert route: %w", err)
	}
	if err := r.replaceConditions(ctx, m.ID, in.Conditions); err != nil {
		return entity.AlertRoute{}, err
	}
	return entity.AlertRoute{ID: m.ID, WorkspaceID: workspaceID, Position: m.Position, PolicyID: m.EscalationPolicyID, PolicySlug: in.PolicySlug, Conditions: in.Conditions}, nil
}

func (r *repo) Update(ctx context.Context, workspaceID, routeID string, in entity.NewAlertRoute) (entity.AlertRoute, error) {
	m, err := r.find(ctx, workspaceID, routeID)
	if err != nil {
		return entity.AlertRoute{}, err
	}
	m.EscalationPolicyID = in.PolicyID
	m.UpdatedAt = time.Now().UTC()
	if _, err := m.Update(ctx, r.db.Querier(ctx), boil.Whitelist("escalation_policy_id", "updated_at")); err != nil {
		return entity.AlertRoute{}, fmt.Errorf("update alert route: %w", err)
	}
	if err := r.replaceConditions(ctx, m.ID, in.Conditions); err != nil {
		return entity.AlertRoute{}, err
	}
	return entity.AlertRoute{ID: m.ID, WorkspaceID: workspaceID, Position: m.Position, PolicyID: m.EscalationPolicyID, PolicySlug: in.PolicySlug, Conditions: in.Conditions}, nil
}

func (r *repo) Delete(ctx context.Context, workspaceID, routeID string) error {
	m, err := r.find(ctx, workspaceID, routeID)
	if err != nil {
		return err
	}
	if _, err := m.Delete(ctx, r.db.Querier(ctx)); err != nil {
		return fmt.Errorf("delete alert route: %w", err)
	}
	return nil
}

func (r *repo) Reorder(ctx context.Context, workspaceID string, orderedIDs []string) error {
	exec := r.db.Querier(ctx)
	for i, id := range orderedIDs {
		if _, err := dbpostgres.AlertRoutes(
			qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
		).UpdateAll(ctx, exec, dbpostgres.M{"position": i, "updated_at": time.Now().UTC()}); err != nil {
			return fmt.Errorf("reorder alert routes: %w", err)
		}
	}
	return nil
}

func (r *repo) find(ctx context.Context, workspaceID, routeID string) (*dbpostgres.AlertRoute, error) {
	m, err := dbpostgres.AlertRoutes(qm.Where("workspace_id = ? AND id = ?", workspaceID, routeID)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrAlertRouteNotFound
		}
		return nil, fmt.Errorf("get alert route: %w", err)
	}
	return m, nil
}

func (r *repo) replaceConditions(ctx context.Context, routeID string, conditions []entity.RouteCondition) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.AlertRouteConditions(qm.Where("route_id = ?", routeID)).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("clear route conditions: %w", err)
	}
	for i, c := range conditions {
		row := &dbpostgres.AlertRouteCondition{
			RouteID:  routeID,
			Field:    c.Field,
			Op:       string(c.Op),
			Value:    c.Value,
			Position: i,
		}
		if err := row.Insert(ctx, exec, boil.Whitelist("route_id", "field", "op", "value", "position")); err != nil {
			return fmt.Errorf("insert route condition: %w", err)
		}
	}
	return nil
}

func (r *repo) ListGroupRules(ctx context.Context, workspaceID string) ([]entity.GroupRule, error) {
	rows, err := dbpostgres.AlertGroupRules(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("position, id"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list group rules: %w", err)
	}
	out := make([]entity.GroupRule, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.GroupRule{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			Fields:      []string(m.Fields),
			Window:      time.Duration(m.WindowSeconds) * time.Second,
			Position:    m.Position,
		})
	}
	return out, nil
}

func (r *repo) ReplaceGroupRules(ctx context.Context, workspaceID string, rules []entity.GroupRule) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.AlertGroupRules(qm.Where("workspace_id = ?", workspaceID)).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("clear group rules: %w", err)
	}
	for i, rule := range rules {
		window := int(rule.Window / time.Second)
		if window <= 0 {
			window = int(entity.GroupWindowDefault / time.Second)
		}
		m := &dbpostgres.AlertGroupRule{
			WorkspaceID:   workspaceID,
			Fields:        types.StringArray(rule.Fields),
			WindowSeconds: window,
			Position:      i,
		}
		if err := m.Insert(ctx, exec, boil.Whitelist("workspace_id", "fields", "window_seconds", "position")); err != nil {
			return fmt.Errorf("insert group rule: %w", err)
		}
	}
	return nil
}

func (r *repo) Settings(ctx context.Context, workspaceID string) (entity.AlertSettings, error) {
	m, err := dbpostgres.FindAlertSetting(ctx, r.db.Querier(ctx), workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AlertSettings{WorkspaceID: workspaceID}, nil
		}
		return entity.AlertSettings{}, fmt.Errorf("get alert settings: %w", err)
	}
	out := entity.AlertSettings{WorkspaceID: m.WorkspaceID, DefaultPolicyID: m.DefaultEscalationPolicyID.String}
	if out.DefaultPolicyID != "" {
		slugs, err := r.policySlugs(ctx, workspaceID)
		if err != nil {
			return entity.AlertSettings{}, err
		}
		out.DefaultPolicySlug = slugs[out.DefaultPolicyID]
	}
	return out, nil
}

func (r *repo) SetDefaultPolicy(ctx context.Context, workspaceID, policyID string) error {
	m := &dbpostgres.AlertSetting{WorkspaceID: workspaceID, UpdatedAt: time.Now().UTC()}
	if policyID != "" {
		m.DefaultEscalationPolicyID = null.StringFrom(policyID)
	}
	if err := m.Upsert(ctx, r.db.Querier(ctx), true,
		[]string{"workspace_id"},
		boil.Whitelist("default_escalation_policy_id", "updated_at"),
		boil.Whitelist("workspace_id", "default_escalation_policy_id")); err != nil {
		return fmt.Errorf("set default policy: %w", err)
	}
	return nil
}
