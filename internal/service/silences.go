package service

//go:generate go tool mockgen -source=silences.go -destination=./silences/silences_mock.go -package=silences

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Silences interface {
	List(ctx context.Context, workspaceSlug string) ([]entity.Silence, error)
	Create(ctx context.Context, workspaceSlug string, in entity.NewSilence) (entity.Silence, error)
	End(ctx context.Context, workspaceSlug, silenceID string) error
}
