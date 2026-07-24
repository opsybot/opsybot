package repository

//go:generate go tool mockgen -source=service.go -destination=./service/service_mock.go -package=service

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Service interface {
	List(ctx context.Context, workspaceID string) ([]entity.Service, error)
	GetByID(ctx context.Context, workspaceID, id string) (entity.Service, error)
	SlugExists(ctx context.Context, workspaceID, slug string) (bool, error)
	ExistingIDs(ctx context.Context, workspaceID string, ids []string) ([]string, error)
	Create(ctx context.Context, s entity.Service) (entity.Service, error)
	Update(ctx context.Context, workspaceID, id, name, teamID, description string) (entity.Service, error)
	Delete(ctx context.Context, workspaceID, id string) error
}
