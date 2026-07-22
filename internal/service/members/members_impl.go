package members

import (
	"context"
	"errors"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	cfg        config.Auth
	tx         repository.Transactor
	lock       repository.Lock
	workspaces repository.Workspace
	members    repository.Member
	users      repository.User
	invites    repository.Invite
	sessions   repository.Session
	policy     repository.Policy
	audit      repository.Audit
	mailer     repository.Mailer
	references service.References
}

func New(
	cfg config.Auth,
	tx repository.Transactor,
	lock repository.Lock,
	workspaces repository.Workspace,
	members repository.Member,
	users repository.User,
	invites repository.Invite,
	sessions repository.Session,
	policy repository.Policy,
	audit repository.Audit,
	mailer repository.Mailer,
	references service.References,
) service.Members {
	return &srv{cfg: cfg, tx: tx, lock: lock, workspaces: workspaces, members: members, users: users,
		invites: invites, sessions: sessions, policy: policy, audit: audit, mailer: mailer, references: references}
}

func (s *srv) authorize(ctx context.Context, workspaceSlug string, obj entity.PolicyObject, act entity.PolicyAction) (entity.Identity, entity.Workspace, error) {
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

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.Member, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	list, err := s.members.ListByWorkspace(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	roles, err := s.policy.RolesByWorkspace(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Role = roles[list[i].UserID]
		list[i].References, err = s.references.ListByUser(ctx, ws.ID, list[i].UserID)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (s *srv) Get(ctx context.Context, workspaceSlug, userID string) (entity.Member, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionRead)
	if err != nil {
		return entity.Member{}, err
	}
	m, err := s.members.Get(ctx, ws.ID, userID)
	if err != nil {
		return entity.Member{}, err
	}
	role, _, err := s.policy.RoleOf(ctx, userID, ws.ID)
	if err != nil {
		return entity.Member{}, err
	}
	m.Role = role
	m.References, err = s.references.ListByUser(ctx, ws.ID, userID)
	if err != nil {
		return entity.Member{}, err
	}
	return m, nil
}

func (s *srv) References(ctx context.Context, workspaceSlug, userID string) ([]entity.MemberReference, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionWrite)
	if err != nil {
		return nil, err
	}
	return s.references.ListByUser(ctx, ws.ID, userID)
}

func (s *srv) Invite(ctx context.Context, workspaceSlug, email string, role entity.Role) (entity.Invite, string, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionWrite)
	if err != nil {
		return entity.Invite{}, "", err
	}
	if err := entity.ValidateEmail(email); err != nil {
		return entity.Invite{}, "", err
	}
	if err := role.Validate(); err != nil {
		return entity.Invite{}, "", err
	}

	token, err := entity.GenerateToken(entity.InviteTokenLength)
	if err != nil {
		return entity.Invite{}, "", err
	}

	var inv entity.Invite
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		user, err := s.resolveInvitee(ctx, ws.ID, email)
		if err != nil {
			return err
		}
		if err := s.members.Create(ctx, ws.ID, user.ID, entity.MemberStatusInvited); err != nil {
			return err
		}
		inv, err = s.invites.Create(ctx, ws.ID, user.ID, actor.UserID, entity.HashToken(token), time.Now().Add(s.cfg.InviteTTL))
		if err != nil {
			return err
		}
		if err := s.policy.AssignRole(ctx, user.ID, ws.ID, role); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionMemberInvited, email))
	})
	if err != nil {
		s.compensateRole(ctx, err, inv.UserID, ws.ID)
		return entity.Invite{}, "", err
	}
	inv.Role = role
	acceptURL := s.acceptURL(token)
	if mailErr := s.mailer.SendInvite(ctx, email, actor.Label, ws.Name, acceptURL); mailErr != nil {
		logger.From(ctx).WarnContext(ctx, "invite email failed", "error", mailErr, "workspace_id", ws.ID)
	}
	return inv, acceptURL, nil
}

func (s *srv) resolveInvitee(ctx context.Context, workspaceID, email string) (entity.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return s.users.CreateInvited(ctx, email)
		}
		return entity.User{}, err
	}
	m, err := s.members.Get(ctx, workspaceID, user.ID)
	if err == nil {
		switch m.Status {
		case entity.MemberStatusActive:
			return entity.User{}, entity.ErrMemberAlreadyExists
		case entity.MemberStatusDeactivated:
			return entity.User{}, entity.ErrMemberDeactivated
		default:
			return entity.User{}, entity.ErrInvitePending
		}
	}
	if !errors.Is(err, entity.ErrMemberNotFound) {
		return entity.User{}, err
	}
	return user, nil
}

func (s *srv) ListInvites(ctx context.Context, workspaceSlug string) ([]entity.Invite, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionWrite)
	if err != nil {
		return nil, err
	}
	invites, err := s.invites.ListPending(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	roles, err := s.policy.RolesByWorkspace(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	for i := range invites {
		invites[i].Role = roles[invites[i].UserID]
	}
	return invites, nil
}

func (s *srv) ResendInvite(ctx context.Context, workspaceSlug, userID string) (entity.Invite, string, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionWrite)
	if err != nil {
		return entity.Invite{}, "", err
	}
	inv, err := s.invites.GetPending(ctx, ws.ID, userID)
	if err != nil {
		return entity.Invite{}, "", err
	}
	token, err := entity.GenerateToken(entity.InviteTokenLength)
	if err != nil {
		return entity.Invite{}, "", err
	}
	if err := s.invites.RotateToken(ctx, inv.ID, entity.HashToken(token), time.Now().Add(s.cfg.InviteTTL)); err != nil {
		return entity.Invite{}, "", err
	}
	acceptURL := s.acceptURL(token)
	if mailErr := s.mailer.SendInvite(ctx, inv.Email, actor.Label, ws.Name, acceptURL); mailErr != nil {
		logger.From(ctx).WarnContext(ctx, "invite resend email failed", "error", mailErr, "workspace_id", ws.ID)
	}
	return inv, acceptURL, nil
}

func (s *srv) RevokeInvite(ctx context.Context, workspaceSlug, userID string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		inv, err := s.invites.GetPending(ctx, ws.ID, userID)
		if err != nil {
			return err
		}
		if err := s.invites.MarkRevoked(ctx, inv.ID); err != nil {
			return err
		}
		if err := s.members.DeleteInvited(ctx, ws.ID, userID); err != nil {
			return err
		}
		if err := s.policy.RemoveRole(ctx, userID, ws.ID); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionInviteRevoked, inv.Email))
	})
}

func (s *srv) ChangeRole(ctx context.Context, workspaceSlug, userID string, role entity.Role) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	if err := role.Validate(); err != nil {
		return err
	}
	var restore entity.Role
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		m, err := s.members.Get(ctx, ws.ID, userID)
		if err != nil {
			return err
		}
		if m.Status == entity.MemberStatusDeactivated {
			return entity.ErrMemberDeactivated
		}
		cur, ok, err := s.policy.RoleOfTx(ctx, userID, ws.ID)
		if err != nil {
			return err
		}
		if !ok || cur == role {
			return nil
		}
		if cur == entity.RoleAdmin && role != entity.RoleAdmin {
			admins, err := s.policy.CountActiveAdminsTx(ctx, ws.ID)
			if err != nil {
				return err
			}
			if admins <= 1 {
				return entity.ErrMemberLastAdmin
			}
		}
		if err := s.policy.ReplaceRole(ctx, userID, ws.ID, cur, role); err != nil {
			return err
		}
		restore = cur
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionMemberRoleChange, m.Name+" → "+string(role)))
	})
	if err != nil && restore != "" {
		if compErr := s.policy.ReplaceRole(context.WithoutCancel(ctx), userID, ws.ID, role, restore); compErr != nil {
			logger.From(ctx).ErrorContext(ctx, "role change compensation failed", "error", compErr, "user_id", userID, "workspace_id", ws.ID)
		}
	}
	return err
}

func (s *srv) Deactivate(ctx context.Context, workspaceSlug, userID string, replacements map[string]string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		m, err := s.members.Get(ctx, ws.ID, userID)
		if err != nil {
			return err
		}
		if m.Status != entity.MemberStatusActive {
			return entity.ErrMemberNotDeactivated
		}
		cur, _, err := s.policy.RoleOfTx(ctx, userID, ws.ID)
		if err != nil {
			return err
		}
		if cur == entity.RoleAdmin {
			admins, err := s.policy.CountActiveAdminsTx(ctx, ws.ID)
			if err != nil {
				return err
			}
			if admins <= 1 {
				return entity.ErrMemberLastAdmin
			}
		}
		if err := s.validateReplacements(ctx, ws.ID, userID, replacements); err != nil {
			return err
		}
		refs, err := s.references.ListByUser(ctx, ws.ID, userID)
		if err != nil {
			return err
		}
		if err := s.references.ReassignAll(ctx, ws.ID, userID, replacements); err != nil {
			return err
		}
		for _, ref := range refs {
			if err := s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionReferenceReassign, ref.Label)); err != nil {
				return err
			}
		}
		if err := s.members.UpdateStatus(ctx, ws.ID, userID, entity.MemberStatusDeactivated); err != nil {
			return err
		}
		if remaining, err := s.members.CountOtherActive(ctx, userID, ws.ID); err != nil {
			return err
		} else if remaining == 0 {
			if err := s.sessions.DeleteByUser(ctx, userID); err != nil {
				return err
			}
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionMemberDeactivated, m.Name))
	})
}

func (s *srv) validateReplacements(ctx context.Context, workspaceID, userID string, replacements map[string]string) error {
	for _, toUserID := range replacements {
		if toUserID == userID {
			return entity.ErrMemberReplacementInvalid
		}
		active, err := s.members.IsActive(ctx, workspaceID, toUserID)
		if err != nil {
			return err
		}
		if !active {
			return entity.ErrMemberReplacementInvalid
		}
	}
	return nil
}

func (s *srv) Reactivate(ctx context.Context, workspaceSlug, userID string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyObjectMembers, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Workspace(ctx, ws.ID); err != nil {
			return err
		}
		m, err := s.members.Get(ctx, ws.ID, userID)
		if err != nil {
			return err
		}
		if m.Status != entity.MemberStatusDeactivated {
			return entity.ErrMemberNotDeactivated
		}
		if err := s.members.UpdateStatus(ctx, ws.ID, userID, entity.MemberStatusActive); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionMemberReactivated, m.Name))
	})
}

func (s *srv) compensateRole(ctx context.Context, cause error, userID, workspaceID string) {
	if userID == "" {
		return
	}
	if compErr := s.policy.RemoveRole(context.WithoutCancel(ctx), userID, workspaceID); compErr != nil {
		logger.From(ctx).ErrorContext(ctx, "invite casbin compensation failed", "error", compErr, "cause", cause, "user_id", userID, "workspace_id", workspaceID)
	}
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

func (s *srv) acceptURL(token string) string {
	return s.cfg.BaseURL + "/invite?token=" + token
}
