package repository

//go:generate go tool mockgen -source=workspace.go -destination=./workspace/workspace_mock.go -package=workspace

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Workspace interface {
	Create(ctx context.Context, w entity.NewWorkspace, createdBy string) (entity.Workspace, error)
	GetByID(ctx context.Context, id string) (entity.Workspace, error)
	GetBySlug(ctx context.Context, slug string) (entity.Workspace, error)
	Update(ctx context.Context, id string, u entity.WorkspaceUpdate) error
	ListActiveByUser(ctx context.Context, userID string) ([]entity.Workspace, error)
}
