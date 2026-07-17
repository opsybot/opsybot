package service

//go:generate go tool mockgen -source=teams.go -destination=./teams/teams_mock.go -package=teams

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Teams interface {
	List(ctx context.Context, workspaceSlug string, includeArchived bool) ([]entity.Team, error)
	Get(ctx context.Context, workspaceSlug, teamSlug string) (entity.Team, error)
	Create(ctx context.Context, workspaceSlug string, in entity.NewTeam) (entity.Team, error)
	Update(ctx context.Context, workspaceSlug, teamSlug string, in entity.TeamUpdate) (entity.Team, error)
	Archive(ctx context.Context, workspaceSlug, teamSlug string) (entity.Team, error)
	Unarchive(ctx context.Context, workspaceSlug, teamSlug string) (entity.Team, error)
}
