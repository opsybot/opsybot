package notification_rules

import (
	"context"
	"errors"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx         repository.Transactor
	workspaces repository.Workspace
	members    repository.Member
	users      repository.User
	rules      repository.NotificationRule
	channels   repository.Channel
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	tx repository.Transactor,
	workspaces repository.Workspace,
	members repository.Member,
	users repository.User,
	rules repository.NotificationRule,
	channels repository.Channel,
	policy repository.Policy,
	audit repository.Audit,
) service.NotificationRules {
	return &srv{tx: tx, workspaces: workspaces, members: members, users: users, rules: rules, channels: channels, policy: policy, audit: audit}
}

func (s *srv) authorize(ctx context.Context, workspaceSlug string, act entity.PolicyAction) (entity.Identity, entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok || id.Kind != entity.IdentityKindSession {
		return entity.Identity{}, entity.Workspace{}, entity.ErrUnauthenticated
	}
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	active, err := s.members.IsActive(ctx, ws.ID, id.UserID)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !active {
		return entity.Identity{}, entity.Workspace{}, entity.ErrNotMember
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectChannels, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) Get(ctx context.Context, workspaceSlug string) (entity.NotificationSettings, error) {
	id, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.NotificationSettings{}, err
	}
	rule, err := s.rules.Get(ctx, ws.ID, id.UserID)
	if err != nil {
		if errors.Is(err, entity.ErrNotificationRuleNotFound) {
			rule = entity.DefaultNotificationRule(ws.ID, id.UserID)
		} else {
			return entity.NotificationSettings{}, err
		}
	}
	channels, err := s.channels.ListByUser(ctx, id.UserID)
	if err != nil {
		return entity.NotificationSettings{}, err
	}
	return entity.NotificationSettings{Rule: rule, Channels: channels}, nil
}

func (s *srv) Save(ctx context.Context, workspaceSlug string, in entity.NotificationRule) (entity.NotificationRule, error) {
	id, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.NotificationRule{}, err
	}
	in.WorkspaceID = ws.ID
	in.UserID = id.UserID
	if in.QuietHours.Enabled && in.QuietHours.Window.Timezone == "" {
		user, err := s.users.GetByID(ctx, id.UserID)
		if err != nil {
			return entity.NotificationRule{}, err
		}
		in.QuietHours.Window.Timezone = user.Timezone
	}
	if err := in.Validate(); err != nil {
		return entity.NotificationRule{}, err
	}
	channels, err := s.channels.ListByUser(ctx, id.UserID)
	if err != nil {
		return entity.NotificationRule{}, err
	}
	connected := connectedSet(channels)
	if err := entity.ValidateNotificationReach(in, connected); err != nil {
		return entity.NotificationRule{}, err
	}
	var saved entity.NotificationRule
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		saved, err = s.rules.Save(ctx, in)
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, entity.AuditEvent{
			WorkspaceID: ws.ID, ActorType: entity.AuditActorUser, ActorUserID: id.UserID,
			ActorLabel: id.Label, Action: entity.ActionNotificationRulesSaved, Target: "notification rules",
		})
	})
	return saved, err
}

func connectedSet(channels []entity.Channel) func(entity.ChannelType) bool {
	seen := make(map[entity.ChannelType]struct{}, len(channels))
	for _, ch := range channels {
		seen[ch.Type] = struct{}{}
	}
	return func(t entity.ChannelType) bool {
		if t == entity.ChannelTypeEmail {
			return true
		}
		_, ok := seen[t]
		return ok
	}
}
