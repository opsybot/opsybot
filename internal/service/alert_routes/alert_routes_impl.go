package alert_routes

import (
	"context"
	"fmt"
	"strings"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	workspaces repository.Workspace
	members    repository.Member
	routes     repository.AlertRoute
	policies   repository.EscalationPolicy
	policy     repository.Policy
}

func New(workspaces repository.Workspace, members repository.Member, routes repository.AlertRoute, policies repository.EscalationPolicy, policy repository.Policy) service.AlertRoutes {
	return &srv{workspaces: workspaces, members: members, routes: routes, policies: policies, policy: policy}
}

func (s *srv) resolvePolicy(ctx context.Context, workspaceID string, in *entity.NewAlertRoute) error {
	in.PolicySlug = strings.TrimSpace(in.PolicySlug)
	if err := in.Validate(); err != nil {
		return err
	}
	p, err := s.policies.GetBySlug(ctx, workspaceID, in.PolicySlug)
	if err != nil {
		return err
	}
	in.PolicyID = p.ID
	return nil
}

func (s *srv) authorize(ctx context.Context, workspaceSlug string, act entity.PolicyAction, obj entity.PolicyObject) (entity.Identity, entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.Identity{}, entity.Workspace{}, entity.ErrUnauthenticated
	}
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if id.Kind == entity.IdentityKindAPIKey && id.WorkspaceID != ws.ID {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	if id.UserID != "" {
		active, err := s.members.IsActive(ctx, ws.ID, id.UserID)
		if err != nil {
			return entity.Identity{}, entity.Workspace{}, err
		}
		if !active {
			return entity.Identity{}, entity.Workspace{}, entity.ErrNotMember
		}
	}
	if !id.ScopePermits(obj, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, obj, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.AlertRoute, entity.AlertSettings, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectAlertSources)
	if err != nil {
		return nil, entity.AlertSettings{}, err
	}
	routes, err := s.routes.List(ctx, ws.ID)
	if err != nil {
		return nil, entity.AlertSettings{}, err
	}
	settings, err := s.routes.Settings(ctx, ws.ID)
	if err != nil {
		return nil, entity.AlertSettings{}, err
	}
	return routes, settings, nil
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.NewAlertRoute) (entity.AlertRoute, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlertSources)
	if err != nil {
		return entity.AlertRoute{}, err
	}
	if err := s.resolvePolicy(ctx, ws.ID, &in); err != nil {
		return entity.AlertRoute{}, err
	}
	created, err := s.routes.Create(ctx, ws.ID, in)
	if err != nil {
		return entity.AlertRoute{}, fmt.Errorf("create alert route: %w", err)
	}
	return created, nil
}

func (s *srv) Update(ctx context.Context, workspaceSlug, routeID string, in entity.NewAlertRoute) (entity.AlertRoute, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlertSources)
	if err != nil {
		return entity.AlertRoute{}, err
	}
	if err := s.resolvePolicy(ctx, ws.ID, &in); err != nil {
		return entity.AlertRoute{}, err
	}
	return s.routes.Update(ctx, ws.ID, routeID, in)
}

func (s *srv) Delete(ctx context.Context, workspaceSlug, routeID string) error {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlertSources)
	if err != nil {
		return err
	}
	return s.routes.Delete(ctx, ws.ID, routeID)
}

func (s *srv) Reorder(ctx context.Context, workspaceSlug string, ids []string) error {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlertSources)
	if err != nil {
		return err
	}
	return s.routes.Reorder(ctx, ws.ID, ids)
}

func (s *srv) Preview(ctx context.Context, workspaceSlug, payload string) (entity.RoutePreview, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectAlertSources)
	if err != nil {
		return entity.RoutePreview{}, err
	}
	alert, err := entity.PreviewAlert(payload)
	if err != nil {
		return entity.RoutePreview{}, err
	}

	routes, err := s.routes.List(ctx, ws.ID)
	if err != nil {
		return entity.RoutePreview{}, err
	}
	settings, err := s.routes.Settings(ctx, ws.ID)
	if err != nil {
		return entity.RoutePreview{}, err
	}
	groupRules, err := s.routes.ListGroupRules(ctx, ws.ID)
	if err != nil {
		return entity.RoutePreview{}, err
	}

	matched, policyID, hit := entity.RouteFor(routes, alert, settings.DefaultPolicyID)
	out := entity.RoutePreview{Position: -1}
	if hit {
		out.MatchedRouteID = matched.ID
		out.Position = matched.Position
		out.PolicySlug = matched.PolicySlug
	} else if policyID != "" {
		out.PolicySlug = settings.DefaultPolicySlug
	}
	if rule, _, grouped := entity.GroupKeyFor(groupRules, alert); grouped {
		out.GroupFields = rule.Fields
	}
	return out, nil
}

func (s *srv) ListGroupRules(ctx context.Context, workspaceSlug string) ([]entity.GroupRule, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectAlertSources)
	if err != nil {
		return nil, err
	}
	return s.routes.ListGroupRules(ctx, ws.ID)
}

func (s *srv) SaveGroupRules(ctx context.Context, workspaceSlug string, rules []entity.GroupRule) ([]entity.GroupRule, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlertSources)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		rules[i].Position = i
		if rules[i].Window == 0 {
			rules[i].Window = entity.GroupWindowDefault
		}
	}
	if err := entity.ValidateGroupRules(rules); err != nil {
		return nil, err
	}
	if err := s.routes.ReplaceGroupRules(ctx, ws.ID, rules); err != nil {
		return nil, fmt.Errorf("save alert group rules: %w", err)
	}
	return s.routes.ListGroupRules(ctx, ws.ID)
}

func (s *srv) SetDefaultPolicy(ctx context.Context, workspaceSlug, policyRef string) error {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlertSources)
	if err != nil {
		return err
	}
	ref := strings.TrimSpace(policyRef)
	if err := entity.ValidatePolicyRef(ref); err != nil {
		return err
	}
	p, err := s.policies.GetBySlug(ctx, ws.ID, ref)
	if err != nil {
		return err
	}
	return s.routes.SetDefaultPolicy(ctx, ws.ID, p.ID)
}
