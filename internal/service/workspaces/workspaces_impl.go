package workspaces

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

func New(workspaces repository.Workspace, members repository.Member, policy repository.Policy) service.Workspaces {
	return &srv{workspaces: workspaces, members: members, policy: policy}
}

func (s *srv) List(ctx context.Context) ([]entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return nil, entity.ErrUnauthenticated
	}
	if id.UserID == "" {
		if id.WorkspaceID == "" {
			return []entity.Workspace{}, nil
		}
		ws, err := s.workspaces.GetByID(ctx, id.WorkspaceID)
		if err != nil {
			return nil, err
		}
		return []entity.Workspace{ws}, nil
	}
	list, err := s.workspaces.ListActiveByUser(ctx, id.UserID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		role, _, err := s.policy.RoleOf(ctx, id.UserID, list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].Role = role
	}
	return list, nil
}

func (s *srv) Get(ctx context.Context, slug string) (entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.Workspace{}, entity.ErrUnauthenticated
	}
	ws, err := s.workspaces.GetBySlug(ctx, slug)
	if err != nil {
		return entity.Workspace{}, err
	}
	if id.UserID == "" {
		if id.Kind == entity.IdentityKindAPIKey && id.WorkspaceID == ws.ID {
			return ws, nil
		}
		return entity.Workspace{}, entity.ErrNotMember
	}
	active, err := s.members.IsActive(ctx, ws.ID, id.UserID)
	if err != nil {
		return entity.Workspace{}, err
	}
	if !active {
		return entity.Workspace{}, entity.ErrNotMember
	}
	role, _, err := s.policy.RoleOf(ctx, id.UserID, ws.ID)
	if err != nil {
		return entity.Workspace{}, err
	}
	ws.Role = role
	return ws, nil
}
