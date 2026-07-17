package service

//go:generate go tool mockgen -source=workspaces.go -destination=./workspaces/workspaces_mock.go -package=workspaces

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Workspaces interface {
	List(ctx context.Context) ([]entity.Workspace, error)
	Get(ctx context.Context, slug string) (entity.Workspace, error)
}
