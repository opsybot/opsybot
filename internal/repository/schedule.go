package repository

//go:generate go tool mockgen -source=schedule.go -destination=./schedule/schedule_mock.go -package=schedule

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Schedule interface {
	Create(ctx context.Context, s entity.Schedule) (entity.Schedule, error)
	GetBySlug(ctx context.Context, workspaceID, slug string) (entity.Schedule, error)
	GetByID(ctx context.Context, workspaceID, id string) (entity.Schedule, error)
	GetByFeedToken(ctx context.Context, feedToken string) (entity.Schedule, error)
	ListByWorkspace(ctx context.Context, workspaceID string, includeArchived bool) ([]entity.Schedule, error)
	ListActive(ctx context.Context, workspaceID string) ([]entity.Schedule, error)
	Update(ctx context.Context, workspaceID, slug string, s entity.Schedule) (entity.Schedule, error)
	SetArchived(ctx context.Context, workspaceID, slug string, archived bool) (entity.Schedule, error)
	SetPaused(ctx context.Context, workspaceID, slug string, paused bool) (entity.Schedule, error)
	Delete(ctx context.Context, workspaceID, slug string) error
	SlugExists(ctx context.Context, workspaceID, slug string) (bool, error)
	AddOverride(ctx context.Context, workspaceID, scheduleID string, o entity.Override) (entity.Override, error)
	ListReferencesByUser(ctx context.Context, workspaceID, userID string) ([]entity.MemberReference, error)
	Reassign(ctx context.Context, workspaceID, scheduleID, fromUserID, toUserID string) error
}
