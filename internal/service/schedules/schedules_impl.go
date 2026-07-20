package schedules

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx         repository.Transactor
	lock       repository.Lock
	workspaces repository.Workspace
	members    repository.Member
	teams      repository.Team
	schedules  repository.Schedule
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	tx repository.Transactor,
	lock repository.Lock,
	workspaces repository.Workspace,
	members repository.Member,
	teams repository.Team,
	schedules repository.Schedule,
	policy repository.Policy,
	audit repository.Audit,
) service.Schedules {
	return &srv{tx: tx, lock: lock, workspaces: workspaces, members: members, teams: teams, schedules: schedules, policy: policy, audit: audit}
}

func (s *srv) authorize(ctx context.Context, workspaceSlug string, act entity.PolicyAction) (entity.Identity, entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.Identity{}, entity.Workspace{}, entity.ErrUnauthenticated
	}
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if id.Kind == entity.IdentityKindAPIKey && id.WorkspaceID != ws.ID {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	if id.UserID != "" {
		active, err := s.members.IsActive(ctx, ws.ID, id.UserID)
		if err != nil {
			return entity.Identity{}, entity.Workspace{}, err
		}
		if !active {
			return entity.Identity{}, entity.Workspace{}, entity.ErrNotMember
		}
	}
	if !id.ScopePermits(entity.PolicyObjectSchedules, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectSchedules, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string, includeArchived bool) ([]entity.Schedule, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	return s.schedules.ListByWorkspace(ctx, ws.ID, includeArchived)
}

func (s *srv) Get(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.Schedule{}, err
	}
	return s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.NewSchedule) (entity.Schedule, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Schedule{}, err
	}
	if err := in.Validate(); err != nil {
		return entity.Schedule{}, err
	}
	layers := buildLayers(in.Layers)

	var created entity.Schedule
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		team, err := s.resolveTeam(ctx, ws.ID, in.TeamSlug)
		if err != nil {
			return err
		}
		if err := s.ensureActiveMembers(ctx, ws.ID, layers); err != nil {
			return err
		}
		if slices.Contains(entity.ScheduleReservedSlugs, in.Slug) {
			return entity.ErrScheduleSlugTaken
		}
		exists, err := s.schedules.SlugExists(ctx, ws.ID, in.Slug)
		if err != nil {
			return err
		}
		if exists {
			return entity.ErrScheduleSlugTaken
		}
		token, err := entity.GenerateToken(entity.FeedTokenLength)
		if err != nil {
			return err
		}
		created, err = s.schedules.Create(ctx, entity.Schedule{
			WorkspaceID: ws.ID, TeamID: team.ID, Slug: in.Slug, Timezone: in.Timezone, FeedToken: token, Layers: layers,
		})
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionScheduleCreated, created.Slug))
	})
	if err != nil {
		return entity.Schedule{}, err
	}
	return created, nil
}

func (s *srv) Update(ctx context.Context, workspaceSlug, scheduleSlug string, in entity.ScheduleUpdate) (entity.Schedule, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Schedule{}, err
	}
	if err := in.Validate(); err != nil {
		return entity.Schedule{}, err
	}
	layers := buildLayers(in.Layers)

	var updated entity.Schedule
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		current, err := s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
		if err != nil {
			return err
		}
		if current.Archived {
			return entity.ErrScheduleArchived
		}
		team, err := s.resolveTeam(ctx, ws.ID, in.TeamSlug)
		if err != nil {
			return err
		}
		if err := s.ensureActiveMembers(ctx, ws.ID, layers); err != nil {
			return err
		}
		if in.Slug != scheduleSlug {
			if slices.Contains(entity.ScheduleReservedSlugs, in.Slug) {
				return entity.ErrScheduleSlugTaken
			}
			exists, err := s.schedules.SlugExists(ctx, ws.ID, in.Slug)
			if err != nil {
				return err
			}
			if exists {
				return entity.ErrScheduleSlugTaken
			}
		}
		updated, err = s.schedules.Update(ctx, ws.ID, scheduleSlug, entity.Schedule{
			Slug: in.Slug, TeamID: team.ID, Timezone: in.Timezone, Layers: layers,
		})
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionScheduleUpdated, updated.Slug))
	})
	if err != nil {
		return entity.Schedule{}, err
	}
	return updated, nil
}

func (s *srv) Duplicate(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Schedule{}, err
	}
	var dup entity.Schedule
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		current, err := s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
		if err != nil {
			return err
		}
		slug, err := s.freeSlug(ctx, ws.ID, current.Slug+"-copy")
		if err != nil {
			return err
		}
		token, err := entity.GenerateToken(entity.FeedTokenLength)
		if err != nil {
			return err
		}
		dup, err = s.schedules.Create(ctx, entity.Schedule{
			WorkspaceID: ws.ID, TeamID: current.TeamID, Slug: slug, Timezone: current.Timezone, FeedToken: token, Layers: current.Layers,
		})
		if err != nil {
			return err
		}
		dup, err = s.schedules.SetPaused(ctx, ws.ID, slug, true)
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionScheduleDuplicated, dup.Slug))
	})
	if err != nil {
		return entity.Schedule{}, err
	}
	return dup, nil
}

func (s *srv) Archive(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error) {
	return s.setArchived(ctx, workspaceSlug, scheduleSlug, true)
}

func (s *srv) Unarchive(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error) {
	return s.setArchived(ctx, workspaceSlug, scheduleSlug, false)
}

func (s *srv) setArchived(ctx context.Context, workspaceSlug, scheduleSlug string, archived bool) (entity.Schedule, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Schedule{}, err
	}
	var out entity.Schedule
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		current, err := s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
		if err != nil {
			return err
		}
		if current.Archived == archived {
			if archived {
				return entity.ErrScheduleArchived
			}
			return entity.ErrScheduleNotArchived
		}
		out, err = s.schedules.SetArchived(ctx, ws.ID, scheduleSlug, archived)
		if err != nil {
			return err
		}
		action := entity.ActionScheduleArchived
		if !archived {
			action = entity.ActionScheduleUnarchived
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, action, out.Slug))
	})
	if err != nil {
		return entity.Schedule{}, err
	}
	return out, nil
}

func (s *srv) Pause(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error) {
	return s.setPaused(ctx, workspaceSlug, scheduleSlug, true)
}

func (s *srv) Resume(ctx context.Context, workspaceSlug, scheduleSlug string) (entity.Schedule, error) {
	return s.setPaused(ctx, workspaceSlug, scheduleSlug, false)
}

func (s *srv) setPaused(ctx context.Context, workspaceSlug, scheduleSlug string, paused bool) (entity.Schedule, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Schedule{}, err
	}
	var out entity.Schedule
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		current, err := s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
		if err != nil {
			return err
		}
		if current.Paused == paused {
			if paused {
				return entity.ErrSchedulePaused
			}
			return entity.ErrScheduleNotPaused
		}
		out, err = s.schedules.SetPaused(ctx, ws.ID, scheduleSlug, paused)
		if err != nil {
			return err
		}
		action := entity.ActionSchedulePaused
		if !paused {
			action = entity.ActionScheduleResumed
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, action, out.Slug))
	})
	if err != nil {
		return entity.Schedule{}, err
	}
	return out, nil
}

func (s *srv) Delete(ctx context.Context, workspaceSlug, scheduleSlug string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		current, err := s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
		if err != nil {
			return err
		}
		if !current.Archived {
			return entity.ErrScheduleNotArchived
		}
		if err := s.schedules.Delete(ctx, ws.ID, scheduleSlug); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionScheduleDeleted, current.Slug))
	})
}

func (s *srv) AddOverride(ctx context.Context, workspaceSlug, scheduleSlug string, in entity.NewOverride) (entity.Override, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Override{}, err
	}
	if err := in.Validate(); err != nil {
		return entity.Override{}, err
	}
	if !in.EndsAt.After(in.StartsAt) {
		return entity.Override{}, entity.ErrScheduleOverrideWindow
	}

	var created entity.Override
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		sched, err := s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
		if err != nil {
			return err
		}
		if sched.Archived {
			return entity.ErrScheduleArchived
		}
		active, err := s.members.IsActive(ctx, ws.ID, in.UserID)
		if err != nil {
			return err
		}
		if !active {
			return entity.ErrScheduleParticipant
		}
		if sched.OverrideConflicts(in.StartsAt, in.EndsAt) {
			return entity.ErrScheduleOverrideConflict
		}
		if sched.OverrideRedundant(in.UserID, in.StartsAt, in.EndsAt) {
			return entity.ErrScheduleOverrideNoChange
		}
		created, err = s.schedules.AddOverride(ctx, ws.ID, sched.ID, entity.Override{
			UserID: in.UserID, StartsAt: in.StartsAt, EndsAt: in.EndsAt, Reason: in.Reason, CreatedByUserID: actor.UserID,
		})
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionScheduleOverrideAdd, sched.Slug))
	})
	if err != nil {
		return entity.Override{}, err
	}
	return created, nil
}

func (s *srv) Calendar(ctx context.Context, workspaceSlug, scheduleSlug string, from, to time.Time) (entity.ScheduleCalendar, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.ScheduleCalendar{}, err
	}
	sched, err := s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
	if err != nil {
		return entity.ScheduleCalendar{}, err
	}
	return sched.Calendar(from, to, entity.ScheduleHandoverLimit), nil
}

func (s *srv) OnCall(ctx context.Context, workspaceSlug, scheduleSlug string, at time.Time) (entity.Coverage, time.Time, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.Coverage{}, time.Time{}, err
	}
	sched, err := s.schedules.GetBySlug(ctx, ws.ID, scheduleSlug)
	if err != nil {
		return entity.Coverage{}, time.Time{}, err
	}
	cover := sched.OnCallAt(at)
	var until time.Time
	if seg, ok := sched.OnCallSegment(at); ok {
		until = seg.EndsAt
	}
	return cover, until, nil
}

func (s *srv) Preview(ctx context.Context, workspaceSlug string, in entity.NewSchedule, from, to time.Time) (entity.ScheduleCalendar, error) {
	_, _, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.ScheduleCalendar{}, err
	}
	draft := entity.Schedule{Timezone: in.Timezone, Layers: buildLayers(in.Layers)}
	return draft.Calendar(from, to, entity.ScheduleHandoverLimit), nil
}

func (s *srv) MyShifts(ctx context.Context, workspaceSlug string, from, to time.Time) ([]entity.Shift, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	if actor.UserID == "" {
		return []entity.Shift{}, nil
	}
	scheds, err := s.schedules.ListActive(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	return entity.Shifts(scheds, actor.UserID, from, to), nil
}

func (s *srv) Feed(ctx context.Context, feedToken string) (entity.Schedule, []entity.FeedShift, error) {
	sched, err := s.schedules.GetByFeedToken(ctx, feedToken)
	if err != nil {
		return entity.Schedule{}, nil, err
	}
	if sched.Paused || sched.Archived {
		return sched, []entity.FeedShift{}, nil
	}

	names, err := s.memberNames(ctx, sched.WorkspaceID)
	if err != nil {
		return entity.Schedule{}, nil, err
	}
	now := time.Now()
	from := now.Add(-entity.FeedPastDays * 24 * time.Hour)
	to := now.Add(entity.FeedFutureDays * 24 * time.Hour)

	var shifts []entity.FeedShift
	for _, seg := range sched.Segments(from, to, -1) {
		if seg.UserID == "" {
			continue
		}
		shifts = append(shifts, entity.FeedShift{
			StartsAt: seg.StartsAt, EndsAt: seg.EndsAt, UserID: seg.UserID, UserName: names[seg.UserID],
		})
	}
	return sched, shifts, nil
}

func (s *srv) ListByUser(ctx context.Context, workspaceID, userID string) ([]entity.MemberReference, error) {
	return s.schedules.ListReferencesByUser(ctx, workspaceID, userID)
}

func (s *srv) Reassign(ctx context.Context, workspaceID, referenceID, toUserID string) error {
	scheduleID, fromUserID, ok := strings.Cut(referenceID, ":")
	if !ok {
		return entity.ErrReferenceUnknown
	}
	return s.schedules.Reassign(ctx, workspaceID, scheduleID, fromUserID, toUserID)
}

func (s *srv) resolveTeam(ctx context.Context, workspaceID, teamSlug string) (entity.Team, error) {
	team, err := s.teams.GetBySlug(ctx, workspaceID, teamSlug)
	if err != nil {
		if errors.Is(err, entity.ErrTeamNotFound) {
			return entity.Team{}, entity.ErrScheduleTeamInvalid
		}
		return entity.Team{}, err
	}
	return team, nil
}

func (s *srv) ensureActiveMembers(ctx context.Context, workspaceID string, layers []entity.Layer) error {
	for _, userID := range participantIDs(layers) {
		active, err := s.members.IsActive(ctx, workspaceID, userID)
		if err != nil {
			return err
		}
		if !active {
			return entity.ErrScheduleParticipant
		}
	}
	return nil
}

func (s *srv) freeSlug(ctx context.Context, workspaceID, base string) (string, error) {
	base = entity.Slugify(base)
	for n := 1; n <= entity.ScheduleSlugMaxCandidates; n++ {
		candidate := entity.ScheduleSlugCandidate(base, n)
		if slices.Contains(entity.ScheduleReservedSlugs, candidate) {
			continue
		}
		exists, err := s.schedules.SlugExists(ctx, workspaceID, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", entity.ErrScheduleSlugTaken
}

func (s *srv) memberNames(ctx context.Context, workspaceID string) (map[string]string, error) {
	members, err := s.members.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(members))
	for _, m := range members {
		names[m.UserID] = m.Name
	}
	return names, nil
}

func (s *srv) event(actor entity.Identity, workspaceID, action, target string) entity.AuditEvent {
	return entity.AuditEvent{
		WorkspaceID: workspaceID,
		ActorType:   entity.AuditActorUser,
		ActorUserID: actor.UserID,
		ActorLabel:  actor.Label,
		Action:      action,
		Target:      target,
		IP:          actor.IP,
	}
}

func buildLayers(in []entity.NewScheduleLayer) []entity.Layer {
	out := make([]entity.Layer, 0, len(in))
	for _, l := range in {
		out = append(out, entity.Layer{
			Participants: dedupe(l.Participants),
			Rotation:     l.Rotation,
			IntervalDays: l.IntervalDays,
			HandoverHour: l.HandoverHour,
			StartsOn:     l.StartsOn,
			Restrictions: l.Restrictions,
		})
	}
	return out
}

func participantIDs(layers []entity.Layer) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, l := range layers {
		for _, id := range l.Participants {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
