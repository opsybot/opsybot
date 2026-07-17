package repository

//go:generate go tool mockgen -source=team.go -destination=./team/team_mock.go -package=team

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Team interface {
	Create(ctx context.Context, workspaceID, slug, name string, memberIDs []string) (entity.Team, error)
	GetBySlug(ctx context.Context, workspaceID, slug string) (entity.Team, error)
	ListByWorkspace(ctx context.Context, workspaceID string, includeArchived bool) ([]entity.Team, error)
	Update(ctx context.Context, workspaceID, slug, name string, memberIDs []string) (entity.Team, error)
	SetArchived(ctx context.Context, workspaceID, slug string, archived bool) (entity.Team, error)
	SlugExists(ctx context.Context, workspaceID, slug string) (bool, error)
}
