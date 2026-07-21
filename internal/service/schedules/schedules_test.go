package schedules

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/schedule"
	"github.com/opsybot/opsybot/internal/repository/workspace"
)

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type fakeLock struct{}

func (fakeLock) Workspace(context.Context, string) error { return nil }
func (fakeLock) Instance(context.Context) error          { return nil }

func withUser(userID string) context.Context {
	return entity.WithIdentity(context.Background(), entity.Identity{Kind: entity.IdentityKindSession, UserID: userID})
}

func TestAddOverrideRejectsConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	ws := workspace.NewMockWorkspace(ctrl)
	mem := member.NewMockMember(ctrl)
	pol := policy.NewMockPolicy(ctrl)
	sch := schedule.NewMockSchedule(ctrl)

	ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "w1"}, nil)
	mem.EXPECT().IsActive(gomock.Any(), "w1", "u1").Return(true, nil)
	pol.EXPECT().Allowed(gomock.Any(), "user:u1", "w1", entity.PolicyObjectSchedules, entity.PolicyActionWrite).Return(true, nil)

	existing := entity.Schedule{
		Timezone: "UTC",
		Overrides: []entity.Override{{
			UserID:   "x",
			StartsAt: time.Date(2026, 7, 13, 9, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, 7, 13, 17, 0, 0, 0, time.UTC),
		}},
	}
	sch.EXPECT().GetBySlug(gomock.Any(), "w1", "primary").Return(existing, nil)
	mem.EXPECT().IsActive(gomock.Any(), "w1", "u2").Return(true, nil)

	s := &srv{tx: fakeTx{}, lock: fakeLock{}, workspaces: ws, members: mem, schedules: sch, policy: pol}
	_, err := s.AddOverride(withUser("u1"), "acme", "primary", entity.NewOverride{
		UserID:   "u2",
		StartsAt: time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 7, 13, 20, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, entity.ErrScheduleOverrideConflict) {
		t.Fatalf("got %v, want ErrScheduleOverrideConflict", err)
	}
}

func TestReassignParsesReferenceID(t *testing.T) {
	ctrl := gomock.NewController(t)
	sch := schedule.NewMockSchedule(ctrl)
	sch.EXPECT().Reassign(gomock.Any(), "w1", "sched1", "userFrom", "userTo").Return(nil)

	s := &srv{schedules: sch}
	if err := s.Reassign(context.Background(), "w1", "sched1:userFrom", "userTo"); err != nil {
		t.Fatalf("reassign: %v", err)
	}
}

func TestReassignRejectsMalformedReferenceID(t *testing.T) {
	s := &srv{}
	if err := s.Reassign(context.Background(), "w1", "no-separator", "userTo"); !errors.Is(err, entity.ErrReferenceUnknown) {
		t.Fatalf("got %v, want ErrReferenceUnknown", err)
	}
}

func TestCreateRequiresIdentity(t *testing.T) {
	s := &srv{}
	if _, err := s.Create(context.Background(), "acme", entity.NewSchedule{}); !errors.Is(err, entity.ErrUnauthenticated) {
		t.Fatalf("got %v, want ErrUnauthenticated", err)
	}
}

func TestDeleteRejectsUnarchivedSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	ws := workspace.NewMockWorkspace(ctrl)
	mem := member.NewMockMember(ctrl)
	pol := policy.NewMockPolicy(ctrl)
	sch := schedule.NewMockSchedule(ctrl)

	ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "w1"}, nil)
	mem.EXPECT().IsActive(gomock.Any(), "w1", "u1").Return(true, nil)
	pol.EXPECT().Allowed(gomock.Any(), "user:u1", "w1", entity.PolicyObjectSchedules, entity.PolicyActionWrite).Return(true, nil)
	sch.EXPECT().GetBySlug(gomock.Any(), "w1", "primary").Return(entity.Schedule{Slug: "primary", Archived: false}, nil)

	s := &srv{tx: fakeTx{}, lock: fakeLock{}, workspaces: ws, members: mem, schedules: sch, policy: pol}
	if err := s.Delete(withUser("u1"), "acme", "primary"); !errors.Is(err, entity.ErrScheduleNotArchived) {
		t.Fatalf("got %v, want ErrScheduleNotArchived", err)
	}
}
