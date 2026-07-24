package services

import (
	"context"
	"errors"
	"strings"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx         repository.Transactor
	workspaces repository.Workspace
	members    repository.Member
	teams      repository.Team
	services   repository.Service
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	tx repository.Transactor,
	workspaces repository.Workspace,
	members repository.Member,
	teams repository.Team,
	services repository.Service,
	policy repository.Policy,
	audit repository.Audit,
) service.Services {
	return &srv{tx: tx, workspaces: workspaces, members: members, teams: teams, services: services, policy: policy, audit: audit}
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
	if !id.ScopePermits(entity.PolicyObjectServices, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectServices, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
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

func (s *srv) teamSlugByID(ctx context.Context, workspaceID string) (map[string]string, error) {
	teams, err := s.teams.ListByWorkspace(ctx, workspaceID, true)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(teams))
	for _, t := range teams {
		out[t.ID] = t.Slug
	}
	return out, nil
}

func (s *srv) resolveTeam(ctx context.Context, workspaceID, teamSlug string) (string, error) {
	teamSlug = strings.TrimSpace(teamSlug)
	if teamSlug == "" {
		return "", nil
	}
	team, err := s.teams.GetBySlug(ctx, workspaceID, teamSlug)
	if err != nil {
		if errors.Is(err, entity.ErrTeamNotFound) {
			return "", entity.ErrTeamNotFound
		}
		return "", err
	}
	return team.ID, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.Service, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	list, err := s.services.List(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	slugs, err := s.teamSlugByID(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].TeamSlug = slugs[list[i].TeamID]
	}
	return list, nil
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.NewService) (entity.Service, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Service{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if err := in.Validate(); err != nil {
		return entity.Service{}, err
	}
	teamID, err := s.resolveTeam(ctx, ws.ID, in.TeamSlug)
	if err != nil {
		return entity.Service{}, err
	}
	slug, err := s.freeSlug(ctx, ws.ID, in.Name)
	if err != nil {
		return entity.Service{}, err
	}
	var out entity.Service
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		created, err := s.services.Create(ctx, entity.Service{
			WorkspaceID: ws.ID,
			Slug:        slug,
			Name:        in.Name,
			TeamID:      teamID,
			Description: in.Description,
		})
		if err != nil {
			return err
		}
		out = created
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionServiceCreated, created.Name))
	})
	if err != nil {
		return entity.Service{}, err
	}
	out.TeamSlug = in.TeamSlug
	return out, nil
}

func (s *srv) Update(ctx context.Context, workspaceSlug, id string, in entity.ServiceUpdate) (entity.Service, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.Service{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if err := in.Validate(); err != nil {
		return entity.Service{}, err
	}
	teamID, err := s.resolveTeam(ctx, ws.ID, in.TeamSlug)
	if err != nil {
		return entity.Service{}, err
	}
	var out entity.Service
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		updated, err := s.services.Update(ctx, ws.ID, id, in.Name, teamID, in.Description)
		if err != nil {
			return err
		}
		out = updated
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionServiceUpdated, updated.Name))
	})
	if err != nil {
		return entity.Service{}, err
	}
	out.TeamSlug = in.TeamSlug
	return out, nil
}

func (s *srv) Delete(ctx context.Context, workspaceSlug, id string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		existing, err := s.services.GetByID(ctx, ws.ID, id)
		if err != nil {
			return err
		}
		if err := s.services.Delete(ctx, ws.ID, id); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionServiceDeleted, existing.Name))
	})
}

func (s *srv) freeSlug(ctx context.Context, workspaceID, name string) (string, error) {
	base := entity.Slugify(name)
	for n := 1; n <= entity.ServiceSlugMaxCandidates; n++ {
		candidate := entity.ServiceSlugCandidate(base, n)
		exists, err := s.services.SlugExists(ctx, workspaceID, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", entity.ErrServiceSlugTaken
}
