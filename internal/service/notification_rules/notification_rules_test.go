package notification_rules

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/audit"
	"github.com/opsybot/opsybot/internal/repository/channel"
	"github.com/opsybot/opsybot/internal/repository/chat_identity"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/notification_rule"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/user"
	"github.com/opsybot/opsybot/internal/repository/workspace"
)

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type harness struct {
	srv      *srv
	ws       *workspace.MockWorkspace
	members  *member.MockMember
	users    *user.MockUser
	rules    *notification_rule.MockNotificationRule
	channels *channel.MockChannel
	ids      *chat_identity.MockChatIdentity
	pol      *policy.MockPolicy
	audit    *audit.MockAudit
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	ctrl := gomock.NewController(t)
	h := &harness{
		ws:       workspace.NewMockWorkspace(ctrl),
		members:  member.NewMockMember(ctrl),
		users:    user.NewMockUser(ctrl),
		rules:    notification_rule.NewMockNotificationRule(ctrl),
		channels: channel.NewMockChannel(ctrl),
		ids:      chat_identity.NewMockChatIdentity(ctrl),
		pol:      policy.NewMockPolicy(ctrl),
		audit:    audit.NewMockAudit(ctrl),
	}
	h.srv = &srv{tx: fakeTx{}, workspaces: h.ws, members: h.members, users: h.users, rules: h.rules, channels: h.channels, identities: h.ids, policy: h.pol, audit: h.audit}
	return h
}

func sessionCtx() context.Context {
	return entity.WithIdentity(context.Background(), entity.Identity{Kind: entity.IdentityKindSession, UserID: "u1", Label: "Priya"})
}

func (h *harness) allowSession() {
	h.ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", entity.PolicyObjectChannels, gomock.Any()).Return(true, nil)
}

func TestSaveRejectsNonZeroFirstDelay(t *testing.T) {
	h := newHarness(t)
	h.allowSession()
	h.channels.EXPECT().ListByUser(gomock.Any(), "u1").Return([]entity.Channel{{Type: entity.ChannelTypeEmail}}, nil).AnyTimes()

	_, err := h.srv.Save(sessionCtx(), "acme", entity.NotificationRule{
		High: []entity.NotificationStep{{Channel: entity.ChannelTypeEmail, Delay: 5 * time.Minute}},
	})
	if !entity.IsValidationError(err) {
		t.Fatalf("expected a validation error, got %v", err)
	}
}

func TestSaveRejectsUnconnectedChannel(t *testing.T) {
	h := newHarness(t)
	h.allowSession()
	h.channels.EXPECT().ListByUser(gomock.Any(), "u1").Return([]entity.Channel{{Type: entity.ChannelTypeEmail}}, nil)
	h.ids.EXPECT().LinkedProviders(gomock.Any(), "ws-1", "u1").Return(nil, nil)

	_, err := h.srv.Save(sessionCtx(), "acme", entity.NotificationRule{
		High: []entity.NotificationStep{{Channel: entity.ChannelTypeSlack}},
	})
	if !entity.IsValidationError(err) {
		t.Fatalf("expected a reach validation error, got %v", err)
	}
}

func TestSaveAllowsLinkedChatProvider(t *testing.T) {
	h := newHarness(t)
	h.allowSession()
	h.channels.EXPECT().ListByUser(gomock.Any(), "u1").Return(nil, nil)
	h.ids.EXPECT().LinkedProviders(gomock.Any(), "ws-1", "u1").Return([]entity.ChatProvider{entity.ChatProviderSlack}, nil)
	saved := entity.NotificationRule{WorkspaceID: "ws-1", UserID: "u1", High: []entity.NotificationStep{{Channel: entity.ChannelTypeSlack}}}
	h.rules.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saved, nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	if _, err := h.srv.Save(sessionCtx(), "acme", entity.NotificationRule{
		High: []entity.NotificationStep{{Channel: entity.ChannelTypeSlack}},
	}); err != nil {
		t.Fatalf("save should allow a slack step backed by a linked chat identity: %v", err)
	}
}

func TestSaveAllowsConnectedButUnverifiedChannel(t *testing.T) {
	h := newHarness(t)
	h.allowSession()
	h.channels.EXPECT().ListByUser(gomock.Any(), "u1").Return([]entity.Channel{{Type: entity.ChannelTypeNtfy, Verified: false}}, nil)
	h.ids.EXPECT().LinkedProviders(gomock.Any(), "ws-1", "u1").Return(nil, nil)
	saved := entity.NotificationRule{WorkspaceID: "ws-1", UserID: "u1", High: []entity.NotificationStep{{Channel: entity.ChannelTypeNtfy}}}
	h.rules.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saved, nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	if _, err := h.srv.Save(sessionCtx(), "acme", entity.NotificationRule{
		High: []entity.NotificationStep{{Channel: entity.ChannelTypeNtfy}},
	}); err != nil {
		t.Fatalf("save should allow an unverified but connected channel: %v", err)
	}
}

func TestSaveDefaultsQuietTimezoneToProfile(t *testing.T) {
	h := newHarness(t)
	h.allowSession()
	h.users.EXPECT().GetByID(gomock.Any(), "u1").Return(entity.User{ID: "u1", Timezone: "Europe/Berlin"}, nil)
	h.channels.EXPECT().ListByUser(gomock.Any(), "u1").Return([]entity.Channel{{Type: entity.ChannelTypeEmail}}, nil)
	h.ids.EXPECT().LinkedProviders(gomock.Any(), "ws-1", "u1").Return(nil, nil)
	h.rules.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, rule entity.NotificationRule) (entity.NotificationRule, error) {
			if rule.QuietHours.Window.Timezone != "Europe/Berlin" {
				t.Errorf("quiet tz = %q, want Europe/Berlin", rule.QuietHours.Window.Timezone)
			}
			return rule, nil
		})
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	_, err := h.srv.Save(sessionCtx(), "acme", entity.NotificationRule{
		High: []entity.NotificationStep{{Channel: entity.ChannelTypeEmail}},
		QuietHours: entity.QuietHours{Enabled: true, Window: entity.HoursWindow{
			Days: []int{1, 2, 3}, StartMinute: 22 * 60, EndMinute: 7 * 60,
		}},
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestGetReturnsDefaultWhenNoneStored(t *testing.T) {
	h := newHarness(t)
	h.ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", entity.PolicyObjectChannels, entity.PolicyActionRead).Return(true, nil)
	h.rules.EXPECT().Get(gomock.Any(), "ws-1", "u1").Return(entity.NotificationRule{}, entity.ErrNotificationRuleNotFound)
	h.channels.EXPECT().ListByUser(gomock.Any(), "u1").Return(nil, nil)

	got, err := h.srv.Get(sessionCtx(), "acme")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Rule.High) != 1 || got.Rule.High[0].Channel != entity.ChannelTypeEmail {
		t.Fatalf("default rule = %+v", got.Rule)
	}
}

func TestApiKeyIdentityCannotEditRules(t *testing.T) {
	h := newHarness(t)
	ctx := entity.WithIdentity(context.Background(), entity.Identity{Kind: entity.IdentityKindAPIKey, WorkspaceID: "ws-1"})
	_, err := h.srv.Save(ctx, "acme", entity.NotificationRule{})
	if err != entity.ErrUnauthenticated {
		t.Fatalf("api key should be unauthenticated for rules, got %v", err)
	}
}
