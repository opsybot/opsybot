package members

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	workspaces repository.Workspace
	members    repository.Member
	policy     repository.Policy
}

func New(workspaces repository.Workspace, members repository.Member, policy repository.Policy) service.Members {
	return &srv{workspaces: workspaces, members: members, policy: policy}
}

func (s *srv) authorize(ctx context.Context, workspaceSlug string, obj entity.PolicyObject, act entity.PolicyAction) (entity.Identity, entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.Identity{}, entity.Workspace{}, entity.ErrUnauthenticated
	}
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	active, err := s.members.IsActive(ctx, ws.ID, id.UserID)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !active {
		return entity.Identity{}, entity.Workspace{}, entity.ErrNotMember
	}
	if scope, ok := entity.ScopeFor(obj, act); ok && !id.ScopeAllows(scope) {
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

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.Member, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	list, err := s.members.ListByWorkspace(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	roles, err := s.policy.RolesByWorkspace(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Role = roles[list[i].UserID]
	}
	return list, nil
}

func (s *srv) Get(ctx context.Context, workspaceSlug, userID string) (entity.Member, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionRead)
	if err != nil {
		return entity.Member{}, err
	}
	m, err := s.members.Get(ctx, ws.ID, userID)
	if err != nil {
		return entity.Member{}, err
	}
	role, _, err := s.policy.RoleOf(ctx, userID, ws.ID)
	if err != nil {
		return entity.Member{}, err
	}
	m.Role = role
	return m, nil
}
