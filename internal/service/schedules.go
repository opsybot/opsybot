package service

//go:generate go tool mockgen -source=schedules.go -destination=./schedules/schedules_mock.go -package=schedules

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Schedules interface {
	List(ctx context.Context, workspaceSlug string, includeArchived bool) ([]entity.Schedule, error)
	Get(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error)
	Create(ctx context.Context, workspaceSlug string, in entity.NewSchedule) (entity.Schedule, error)
	Update(ctx context.Context, workspaceSlug, scheduleSlug string, in entity.ScheduleUpdate) (entity.Schedule, error)
	Duplicate(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error)
	Archive(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error)
	Unarchive(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error)
	Pause(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error)
	Resume(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error)
	Delete(ctx context.Context, workspaceSlug, scheduleSlug string) error
	AddOverride(ctx context.Context, workspaceSlug, scheduleSlug string, in entity.NewOverride) (entity.Override, error)
	Calendar(ctx context.Context, workspaceSlug, scheduleSlug string, from, to time.Time) (entity.ScheduleCalendar, error)
	OnCall(ctx context.Context, workspaceSlug, scheduleSlug string, at time.Time) (entity.Coverage, time.Time, error)
	Preview(ctx context.Context, workspaceSlug string, in entity.NewSchedule, from, to time.Time) (entity.ScheduleCalendar, error)
	MyShifts(ctx context.Context, workspaceSlug string, from, to time.Time) ([]entity.Shift, error)
	Feed(ctx context.Context, feedToken string) (entity.Schedule, []entity.FeedShift, error)
}
