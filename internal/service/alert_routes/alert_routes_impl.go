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
	policy     repository.Policy
}

func New(workspaces repository.Workspace, members repository.Member, routes repository.AlertRoute, policy repository.Policy) service.AlertRoutes {
	return &srv{workspaces: workspaces, members: members, routes: routes, policy: policy}
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
	in.PolicyRef = strings.TrimSpace(in.PolicyRef)
	if err := in.Validate(); err != nil {
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
	in.PolicyRef = strings.TrimSpace(in.PolicyRef)
	if err := in.Validate(); err != nil {
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

func (s *srv) SetDefaultPolicy(ctx context.Context, workspaceSlug, policyRef string) error {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlertSources)
	if err != nil {
		return err
	}
	ref := strings.TrimSpace(policyRef)
	if err := entity.ValidatePolicyRef(ref); err != nil {
		return err
	}
	return s.routes.SetDefaultPolicy(ctx, ws.ID, ref)
}
