package alert_monitors

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
	monitors   repository.AlertMonitor
	sources    repository.AlertSource
	routes     repository.AlertRoute
	policies   repository.EscalationPolicy
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	tx repository.Transactor,
	workspaces repository.Workspace,
	members repository.Member,
	monitors repository.AlertMonitor,
	sources repository.AlertSource,
	routes repository.AlertRoute,
	policies repository.EscalationPolicy,
	policy repository.Policy,
	audit repository.Audit,
) service.AlertMonitors {
	return &srv{
		tx: tx, workspaces: workspaces, members: members, monitors: monitors,
		sources: sources, routes: routes, policies: policies, policy: policy, audit: audit,
	}
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

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.AlertMonitor, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	return s.monitors.List(ctx, ws.ID)
}

func (s *srv) Get(ctx context.Context, workspaceSlug, monitorID string) (entity.AlertMonitor, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.AlertMonitor{}, err
	}
	return s.monitors.Get(ctx, ws.ID, monitorID)
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.NewAlertMonitor) (entity.AlertMonitor, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.AlertMonitor{}, err
	}

	in.Name = strings.TrimSpace(in.Name)
	in.Slug = strings.TrimSpace(in.Slug)
	if in.Slug == "" {
		in.Slug = entity.Slugify(in.Name)
	}
	if in.Interval == 0 {
		in.Interval = entity.MonitorIntervalDefault
	}
	if in.Severity == "" {
		in.Severity = entity.SeverityHigh
	}
	in.PolicySlug = strings.TrimSpace(in.PolicySlug)
	if in.PolicySlug == "" {
		settings, err := s.routes.Settings(ctx, ws.ID)
		if err != nil {
			return entity.AlertMonitor{}, err
		}
		in.PolicySlug = settings.DefaultPolicySlug
	}
	if err := in.Validate(); err != nil {
		return entity.AlertMonitor{}, err
	}
	resolved, err := s.policies.GetBySlug(ctx, ws.ID, in.PolicySlug)
	if err != nil {
		return entity.AlertMonitor{}, err
	}
	in.PolicyID = resolved.ID

	token, err := entity.GenerateToken(entity.IngestTokenLength)
	if err != nil {
		return entity.AlertMonitor{}, err
	}
	secret, err := entity.GenerateHexToken(entity.SigningSecretByteLen)
	if err != nil {
		return entity.AlertMonitor{}, err
	}

	var created entity.AlertMonitor
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		source, err := s.sources.Create(ctx, ws.ID, entity.AlertSource{
			Slug:            in.Slug,
			Name:            in.Name,
			Format:          entity.SourceFormatHeartbeat,
			IngestToken:     token,
			SigningSecret:   secret,
			DefaultSeverity: in.Severity,
		})
		if err != nil {
			return fmt.Errorf("create monitor source: %w", err)
		}
		created, err = s.monitors.Create(ctx, ws.ID, source.ID, in)
		if err != nil {
			return fmt.Errorf("create alert monitor: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertMonitorCreated, created.Slug))
	})
	return created, err
}

func (s *srv) Update(ctx context.Context, workspaceSlug, monitorID string, in entity.AlertMonitorUpdate) (entity.AlertMonitor, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.AlertMonitor{}, err
	}

	existing, err := s.monitors.Get(ctx, ws.ID, monitorID)
	if err != nil {
		return entity.AlertMonitor{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Severity == "" {
		in.Severity = existing.Severity
	}
	in.PolicySlug = strings.TrimSpace(in.PolicySlug)
	if in.PolicySlug == "" {
		in.PolicySlug = existing.PolicySlug
	}
	if err := in.Validate(); err != nil {
		return entity.AlertMonitor{}, err
	}
	resolved, err := s.policies.GetBySlug(ctx, ws.ID, in.PolicySlug)
	if err != nil {
		return entity.AlertMonitor{}, err
	}
	in.PolicyID = resolved.ID

	var updated entity.AlertMonitor
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if in.Name != existing.Name {
			if _, err := s.sources.Update(ctx, ws.ID, existing.Slug, entity.AlertSourceUpdate{
				Name:            in.Name,
				DefaultSeverity: in.Severity,
			}); err != nil {
				return fmt.Errorf("update monitor source: %w", err)
			}
		}
		updated, err = s.monitors.Update(ctx, ws.ID, monitorID, in)
		if err != nil {
			return fmt.Errorf("update alert monitor: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertMonitorUpdated, updated.Slug))
	})
	return updated, err
}

func (s *srv) Delete(ctx context.Context, workspaceSlug, monitorID string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	existing, err := s.monitors.Get(ctx, ws.ID, monitorID)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.sources.Delete(ctx, ws.ID, existing.Slug); err != nil {
			return fmt.Errorf("delete monitor source: %w", err)
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertMonitorDeleted, existing.Slug))
	})
}
