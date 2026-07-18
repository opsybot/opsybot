package audits

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	workspaces repository.Workspace
	members    repository.Member
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	workspaces repository.Workspace,
	members repository.Member,
	policy repository.Policy,
	audit repository.Audit,
) service.Audits {
	return &srv{workspaces: workspaces, members: members, policy: policy, audit: audit}
}

func (s *srv) List(ctx context.Context, workspaceSlug string, filter entity.AuditFilter) (entity.AuditPage, error) {
	ws, err := s.authorize(ctx, workspaceSlug)
	if err != nil {
		return entity.AuditPage{}, err
	}
	events, next, err := s.audit.List(ctx, ws.ID, filter)
	if err != nil {
		return entity.AuditPage{}, err
	}
	return entity.AuditPage{Events: events, NextCursor: next}, nil
}

func (s *srv) authorize(ctx context.Context, workspaceSlug string) (entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.Workspace{}, entity.ErrUnauthenticated
	}
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return entity.Workspace{}, err
	}
	if id.Kind == entity.IdentityKindAPIKey && id.WorkspaceID != ws.ID {
		return entity.Workspace{}, entity.ErrForbidden
	}
	if id.UserID != "" {
		active, err := s.members.IsActive(ctx, ws.ID, id.UserID)
		if err != nil {
			return entity.Workspace{}, err
		}
		if !active {
			return entity.Workspace{}, entity.ErrNotMember
		}
	}
	if !id.ScopePermits(entity.PolicyObjectAudit, entity.PolicyActionRead) {
		return entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectAudit, entity.PolicyActionRead)
	if err != nil {
		return entity.Workspace{}, err
	}
	if !allowed {
		return entity.Workspace{}, entity.ErrForbidden
	}
	return ws, nil
}
