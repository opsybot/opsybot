package service

//go:generate go tool mockgen -source=services.go -destination=./services/services_mock.go -package=services

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Services interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.Service, error)
	Create(ctx context.Context, workspaceSlug string, in entity.NewService) (entity.Service, error)
	Update(ctx context.Context, workspaceSlug, id string, in entity.ServiceUpdate) (entity.Service, error)
	Delete(ctx context.Context, workspaceSlug, id string) error
}
