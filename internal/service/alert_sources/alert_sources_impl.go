package alert_sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx         repository.Transactor
	workspaces repository.Workspace
	members    repository.Member
	sources    repository.AlertSource
	events     repository.IngestEvent
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	tx repository.Transactor,
	workspaces repository.Workspace,
	members repository.Member,
	sources repository.AlertSource,
	events repository.IngestEvent,
	policy repository.Policy,
	audit repository.Audit,
) service.AlertSources {
	return &srv{tx: tx, workspaces: workspaces, members: members, sources: sources, events: events, policy: policy, audit: audit}
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
	if !id.ScopePermits(entity.PolicyObjectAlertSources, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectAlertSources, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.AlertSource, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	return s.sources.ListByWorkspace(ctx, ws.ID)
}

func (s *srv) Get(ctx context.Context, workspaceSlug, sourceSlug string) (entity.AlertSource, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.AlertSource{}, err
	}
	return s.sources.GetBySlug(ctx, ws.ID, sourceSlug)
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.NewAlertSource) (entity.AlertSource, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.AlertSource{}, err
	}
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	if in.DefaultSeverity == "" {
		in.DefaultSeverity = entity.SeverityWarning
	}
	if err := in.Validate(); err != nil {
		return entity.AlertSource{}, err
	}

	token, err := entity.GenerateToken(entity.IngestTokenLength)
	if err != nil {
		return entity.AlertSource{}, err
	}
	secret, err := entity.GenerateHexToken(entity.SigningSecretByteLen)
	if err != nil {
		return entity.AlertSource{}, err
	}

	var created entity.AlertSource
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		created, err = s.sources.Create(ctx, ws.ID, entity.AlertSource{
			Slug:             in.Slug,
			Name:             in.Name,
			Format:           in.Format,
			IngestToken:      token,
			SigningSecret:    secret,
			RequireSignature: in.RequireSignature,
			DefaultSeverity:  in.DefaultSeverity,
			AutoResolveAfter: in.AutoResolveAfter,
		})
		if err != nil {
			return fmt.Errorf("create alert source: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertSourceCreated, created.Slug))
	})
	return created, err
}

func (s *srv) Update(ctx context.Context, workspaceSlug, sourceSlug string, in entity.AlertSourceUpdate) (entity.AlertSource, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.AlertSource{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.DefaultSeverity == "" {
		in.DefaultSeverity = entity.SeverityWarning
	}
	if err := in.Validate(); err != nil {
		return entity.AlertSource{}, err
	}

	var updated entity.AlertSource
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		updated, err = s.sources.Update(ctx, ws.ID, sourceSlug, in)
		if err != nil {
			return fmt.Errorf("update alert source: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertSourceUpdated, sourceSlug))
	})
	return updated, err
}

func (s *srv) Delete(ctx context.Context, workspaceSlug, sourceSlug string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.sources.Delete(ctx, ws.ID, sourceSlug); err != nil {
			return fmt.Errorf("delete alert source: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertSourceDeleted, sourceSlug))
	})
}

func (s *srv) SetPaused(ctx context.Context, workspaceSlug, sourceSlug string, paused bool) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	action := entity.ActionAlertSourceResumed
	if paused {
		action = entity.ActionAlertSourcePaused
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.sources.SetPaused(ctx, ws.ID, sourceSlug, paused); err != nil {
			return fmt.Errorf("set alert source paused: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, action, sourceSlug))
	})
}

func (s *srv) RotateSecret(ctx context.Context, workspaceSlug, sourceSlug string) (entity.AlertSource, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.AlertSource{}, err
	}
	secret, err := entity.GenerateHexToken(entity.SigningSecretByteLen)
	if err != nil {
		return entity.AlertSource{}, err
	}

	var rotated entity.AlertSource
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		rotated, err = s.sources.RotateSecret(ctx, ws.ID, sourceSlug, secret)
		if err != nil {
			return fmt.Errorf("rotate alert source secret: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertSourceSecretRotated, sourceSlug))
	})
	return rotated, err
}

func (s *srv) SaveMapping(ctx context.Context, workspaceSlug, sourceSlug string, mappings []entity.SourceMapping) (entity.AlertSource, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.AlertSource{}, err
	}
	src, err := s.sources.GetBySlug(ctx, ws.ID, sourceSlug)
	if err != nil {
		return entity.AlertSource{}, err
	}
	if err := entity.ValidateSourceMappings(src.Format, mappings); err != nil {
		return entity.AlertSource{}, err
	}

	var saved entity.AlertSource
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.sources.ReplaceMappings(ctx, src.ID, mappings); err != nil {
			return fmt.Errorf("replace alert source mappings: %w", err)
		}
		if err := s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertSourceMappingSaved, sourceSlug)); err != nil {
			return err
		}
		saved, err = s.sources.GetBySlug(ctx, ws.ID, sourceSlug)
		return err
	})
	return saved, err
}

func (s *srv) Events(ctx context.Context, workspaceSlug, sourceSlug string, limit int) ([]entity.IngestEvent, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	src, err := s.sources.GetBySlug(ctx, ws.ID, sourceSlug)
	if err != nil {
		return nil, err
	}
	return s.events.ListBySource(ctx, src.ID, limit)
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
