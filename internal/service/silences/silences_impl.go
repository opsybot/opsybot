package silences

import (
	"context"
	"fmt"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	workspaces repository.Workspace
	members    repository.Member
	silences   repository.Silence
	policy     repository.Policy
}

func New(workspaces repository.Workspace, members repository.Member, silences repository.Silence, policy repository.Policy) service.Silences {
	return &srv{workspaces: workspaces, members: members, silences: silences, policy: policy}
}

func (s *srv) authorize(ctx context.Context, workspaceSlug string, act entity.PolicyAction, obj entity.PolicyObject) (entity.Identity, entity.Workspace, error) {
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
	if !id.ScopePermits(obj, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, obj, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.Silence, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectAlerts)
	if err != nil {
		return nil, err
	}
	return s.silences.List(ctx, ws.ID)
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.NewSilence) (entity.Silence, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlerts)
	if err != nil {
		return entity.Silence{}, err
	}
	if in.Kind == "" {
		in.Kind = entity.SilenceKindSilence
	}
	if in.StartsAt.IsZero() {
		in.StartsAt = time.Now().UTC()
	}
	if err := in.Validate(); err != nil {
		return entity.Silence{}, err
	}
	created, err := s.silences.Create(ctx, ws.ID, actor.Label, actor.UserID, in)
	if err != nil {
		return entity.Silence{}, fmt.Errorf("create silence: %w", err)
	}
	return created, nil
}

func (s *srv) End(ctx context.Context, workspaceSlug, silenceID string) error {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectAlerts)
	if err != nil {
		return err
	}
	return s.silences.End(ctx, ws.ID, silenceID, time.Now().UTC())
}
