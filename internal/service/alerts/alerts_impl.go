package alerts

import (
	"context"
	"fmt"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx         repository.Transactor
	workspaces repository.Workspace
	members    repository.Member
	alerts     repository.Alert
	sources    repository.AlertSource
	events     repository.IngestEvent
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	tx repository.Transactor,
	workspaces repository.Workspace,
	members repository.Member,
	alerts repository.Alert,
	sources repository.AlertSource,
	events repository.IngestEvent,
	policy repository.Policy,
	audit repository.Audit,
) service.Alerts {
	return &srv{tx: tx, workspaces: workspaces, members: members, alerts: alerts, sources: sources, events: events, policy: policy, audit: audit}
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
	if !id.ScopePermits(entity.PolicyObjectAlerts, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectAlerts, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string, filter entity.AlertFilter) ([]entity.Alert, string, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, "", err
	}
	list, next, err := s.alerts.List(ctx, ws.ID, filter)
	if err != nil {
		return nil, "", err
	}
	slugs, err := s.sourceSlugs(ctx, ws.ID)
	if err != nil {
		return nil, "", err
	}
	for i := range list {
		list[i].SourceSlug = slugs[list[i].SourceID]
	}
	return list, next, nil
}

func (s *srv) Get(ctx context.Context, workspaceSlug, alertID string) (entity.Alert, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.Alert{}, err
	}
	alert, err := s.alerts.GetByID(ctx, ws.ID, alertID)
	if err != nil {
		return entity.Alert{}, err
	}
	if alert.Timeline, err = s.alerts.ListEvents(ctx, alert.ID, entity.AlertTimelineLimit); err != nil {
		return entity.Alert{}, err
	}
	if alert.Links, err = s.alerts.ListLinks(ctx, alert.ID); err != nil {
		return entity.Alert{}, err
	}
	slugs, err := s.sourceSlugs(ctx, ws.ID)
	if err != nil {
		return entity.Alert{}, err
	}
	alert.SourceSlug = slugs[alert.SourceID]
	return alert, nil
}

func (s *srv) Acknowledge(ctx context.Context, workspaceSlug string, ids []string) (int, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, entity.ErrAlertBulkEmpty
	}

	now := time.Now().UTC()
	var affected int
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		affected, err = s.alerts.Acknowledge(ctx, ws.ID, ids, actor.UserID, actor.Label, now)
		if err != nil {
			return fmt.Errorf("acknowledge alerts: %w", err)
		}
		for _, id := range ids {
			if err := s.alerts.AppendEvent(ctx, id, entity.AlertEvent{
				At:   now,
				Kind: entity.AlertEventAcked,
				Text: fmt.Sprintf("Acknowledged by %s", actor.Label),
			}); err != nil {
				return err
			}
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertAcknowledged, fmt.Sprintf("%d alerts", affected)))
	})
	return affected, err
}

func (s *srv) Resolve(ctx context.Context, workspaceSlug string, ids []string) (int, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, entity.ErrAlertBulkEmpty
	}

	now := time.Now().UTC()
	var affected int
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		affected, err = s.alerts.Resolve(ctx, ws.ID, ids, now, entity.ResolveModeManual)
		if err != nil {
			return fmt.Errorf("resolve alerts: %w", err)
		}
		for _, id := range ids {
			if err := s.alerts.AppendEvent(ctx, id, entity.AlertEvent{
				At:   now,
				Kind: entity.AlertEventResolved,
				Text: fmt.Sprintf("Resolved by %s", actor.Label),
			}); err != nil {
				return err
			}
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionAlertResolved, fmt.Sprintf("%d alerts", affected)))
	})
	return affected, err
}

func (s *srv) Failures(ctx context.Context, workspaceSlug string, limit int) ([]entity.IngestFailure, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	failures, err := s.events.ListFailures(ctx, ws.ID, limit)
	if err != nil {
		return nil, err
	}
	slugs, err := s.sourceSlugs(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	for i := range failures {
		failures[i].SourceSlug = slugs[failures[i].SourceID]
	}
	return failures, nil
}

func (s *srv) sourceSlugs(ctx context.Context, workspaceID string) (map[string]string, error) {
	sources, err := s.sources.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(sources))
	for _, src := range sources {
		out[src.ID] = src.Slug
	}
	return out, nil
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
