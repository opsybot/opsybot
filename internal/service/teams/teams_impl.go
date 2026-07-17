package teams

import (
	"context"
	"slices"

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
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	tx repository.Transactor,
	lock repository.Lock,
	workspaces repository.Workspace,
	members repository.Member,
	teams repository.Team,
	policy repository.Policy,
	audit repository.Audit,
) service.Teams {
	return &srv{tx: tx, lock: lock, workspaces: workspaces, members: members, teams: teams, policy: policy, audit: audit}
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
	if !id.ScopePermits(entity.PolicyObjectTeams, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectTeams, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string, includeArchived bool) ([]entity.Team, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	return s.teams.ListByWorkspace(ctx, ws.ID, includeArchived)
}

func (s *srv) Get(ctx context.Context, workspaceSlug, teamSlug string) (entity.Team, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.Team{}, err
	}
	return s.teams.GetBySlug(ctx, ws.ID, teamSlug)
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.NewTeam) (entity.Team, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Team{}, err
	}
	if err := in.Validate(); err != nil {
		return entity.Team{}, err
	}
	members := dedupe(in.MemberIDs)

	var team entity.Team
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		if err := s.ensureActiveMembers(ctx, ws.ID, members); err != nil {
			return err
		}
		slug, err := s.pickSlug(ctx, ws.ID, in.Name)
		if err != nil {
			return err
		}
		team, err = s.teams.Create(ctx, ws.ID, slug, in.Name, members)
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionTeamCreated, team.Name))
	})
	if err != nil {
		return entity.Team{}, err
	}
	return team, nil
}

func (s *srv) Update(ctx context.Context, workspaceSlug, teamSlug string, in entity.TeamUpdate) (entity.Team, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Team{}, err
	}
	if err := in.Validate(); err != nil {
		return entity.Team{}, err
	}
	members := dedupe(in.MemberIDs)

	var team entity.Team
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		current, err := s.teams.GetBySlug(ctx, ws.ID, teamSlug)
		if err != nil {
			return err
		}
		if current.Archived {
			return entity.ErrTeamArchived
		}
		if err := s.ensureActiveMembers(ctx, ws.ID, members); err != nil {
			return err
		}
		team, err = s.teams.Update(ctx, ws.ID, teamSlug, in.Name, members)
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionTeamUpdated, team.Name))
	})
	if err != nil {
		return entity.Team{}, err
	}
	return team, nil
}

func (s *srv) Archive(ctx context.Context, workspaceSlug, teamSlug string) (entity.Team, error) {
	return s.setArchived(ctx, workspaceSlug, teamSlug, true)
}

func (s *srv) Unarchive(ctx context.Context, workspaceSlug, teamSlug string) (entity.Team, error) {
	return s.setArchived(ctx, workspaceSlug, teamSlug, false)
}

func (s *srv) setArchived(ctx context.Context, workspaceSlug, teamSlug string, archived bool) (entity.Team, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Team{}, err
	}
	var team entity.Team
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		current, err := s.teams.GetBySlug(ctx, ws.ID, teamSlug)
		if err != nil {
			return err
		}
		if current.Archived == archived {
			if archived {
				return entity.ErrTeamArchived
			}
			return entity.ErrTeamNotArchived
		}
		team, err = s.teams.SetArchived(ctx, ws.ID, teamSlug, archived)
		if err != nil {
			return err
		}
		action := entity.ActionTeamArchived
		if !archived {
			action = entity.ActionTeamUnarchived
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, action, team.Name))
	})
	if err != nil {
		return entity.Team{}, err
	}
	return team, nil
}

func (s *srv) ensureActiveMembers(ctx context.Context, workspaceID string, memberIDs []string) error {
	for _, userID := range memberIDs {
		active, err := s.members.IsActive(ctx, workspaceID, userID)
		if err != nil {
			return err
		}
		if !active {
			return entity.ErrTeamMemberInvalid
		}
	}
	return nil
}

func (s *srv) pickSlug(ctx context.Context, workspaceID, name string) (string, error) {
	base := entity.Slugify(name)
	for n := 1; n <= entity.TeamSlugMaxCandidates; n++ {
		candidate := entity.TeamSlugCandidate(base, n)
		if slices.Contains(entity.TeamReservedSlugs, candidate) {
			continue
		}
		exists, err := s.teams.SlugExists(ctx, workspaceID, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", entity.ErrTeamSlugTaken
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
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
