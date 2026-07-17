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
}

func New(workspaces repository.Workspace, members repository.Member) service.Workspaces {
	return &srv{workspaces: workspaces, members: members}
}

func (s *srv) List(ctx context.Context) ([]entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return nil, entity.ErrUnauthenticated
	}
	return s.workspaces.ListActiveByUser(ctx, id.UserID)
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
	active, err := s.members.IsActive(ctx, ws.ID, id.UserID)
	if err != nil {
		return entity.Workspace{}, err
	}
	if !active {
		return entity.Workspace{}, entity.ErrNotMember
	}
	return ws, nil
}
