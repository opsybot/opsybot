package repository

//go:generate go tool mockgen -source=silence.go -destination=./silence/silence_mock.go -package=silence

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Silence interface {
	List(ctx context.Context, workspaceID string) ([]entity.Silence, error)
	ListActive(ctx context.Context, workspaceID string, at time.Time) ([]entity.Silence, error)
	Create(ctx context.Context, workspaceID, createdBy, createdByUserID string, in entity.NewSilence) (entity.Silence, error)
	End(ctx context.Context, workspaceID, silenceID string, at time.Time) error
}
