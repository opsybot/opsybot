package chats

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/audit"
	"github.com/opsybot/opsybot/internal/repository/chat_connection"
	"github.com/opsybot/opsybot/internal/repository/chat_courier"
	"github.com/opsybot/opsybot/internal/repository/chat_identity"
	"github.com/opsybot/opsybot/internal/repository/chat_oauth_state"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/workspace"
)

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type harness struct {
	srv         *srv
	ws          *workspace.MockWorkspace
	members     *member.MockMember
	pol         *policy.MockPolicy
	connections *chat_connection.MockChatConnection
	identities  *chat_identity.MockChatIdentity
	courier     *chat_courier.MockChatCourier
	oauthStates *chat_oauth_state.MockChatOAuthState
	audit       *audit.MockAudit
}

func newHarness(t *testing.T, slack config.Slack, discord config.Discord) *harness {
	t.Helper()
	ctrl := gomock.NewController(t)
	h := &harness{
		ws:          workspace.NewMockWorkspace(ctrl),
		members:     member.NewMockMember(ctrl),
		pol:         policy.NewMockPolicy(ctrl),
		connections: chat_connection.NewMockChatConnection(ctrl),
		identities:  chat_identity.NewMockChatIdentity(ctrl),
		courier:     chat_courier.NewMockChatCourier(ctrl),
		oauthStates: chat_oauth_state.NewMockChatOAuthState(ctrl),
		audit:       audit.NewMockAudit(ctrl),
	}
	h.srv = &srv{
		tx: fakeTx{}, workspaces: h.ws, members: h.members, policy: h.pol,
		connections: h.connections, identities: h.identities, courier: h.courier,
		oauthStates: h.oauthStates, audit: h.audit,
		cfg: config.Auth{BaseURL: "https://opsy.test"}, slack: slack, discord: discord,
	}
	return h
}

func sessionCtx() context.Context {
	return entity.WithIdentity(context.Background(), entity.Identity{Kind: entity.IdentityKindSession, UserID: "u1", Label: "Priya"})
}

func (h *harness) allowWrite() {
	h.ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", entity.PolicyObjectChat, entity.PolicyActionWrite).Return(true, nil)
}

func (h *harness) allowRead() {
	h.ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "ws-1", Slug: "acme"}, nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", entity.PolicyObjectChat, entity.PolicyActionRead).Return(true, nil)
}

func TestConnectUsesEnvTokenAndStoresNoSecret(t *testing.T) {
	h := newHarness(t, config.Slack{BotToken: "xoxb-env-token"}, config.Discord{})
	h.allowWrite()

	h.courier.EXPECT().
		Validate(gomock.Any(), entity.ChatProviderSlack, "xoxb-env-token", "").
		Return(entity.ChatValidation{ExternalID: "T1", ExternalName: "Acme", BotUserID: "U0"}, nil)

	var saved entity.ChatConnectionInput
	h.connections.EXPECT().
		Save(gomock.Any(), "ws-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, in entity.ChatConnectionInput) (entity.ChatConnection, error) {
			saved = in
			return entity.ChatConnection{Provider: in.Provider, ExternalName: in.ExternalName}, nil
		})
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	conn, err := h.srv.Connect(sessionCtx(), "acme", entity.ChatConnectInput{Provider: entity.ChatProviderSlack})
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if conn.ExternalName != "Acme" {
		t.Errorf("ExternalName = %q, want Acme", conn.ExternalName)
	}
	if saved.BotToken != "" {
		t.Errorf("BotToken persisted = %q, want empty: an app-level env token must never be written to the DB", saved.BotToken)
	}
	if saved.ExternalID != "T1" || saved.BotUserID != "U0" {
		t.Errorf("validated identity not carried into Save: %+v", saved)
	}
}

func TestConnectRejectsWhenNoTokenAvailable(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.allowWrite()

	_, err := h.srv.Connect(sessionCtx(), "acme", entity.ChatConnectInput{Provider: entity.ChatProviderSlack})
	if err != entity.ErrChatProviderNotConfigured {
		t.Fatalf("err = %v, want ErrChatProviderNotConfigured when neither UI nor env supplies a token", err)
	}
}

func TestStartOAuthStoresStateAndReturnsURL(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.allowWrite()
	h.connections.EXPECT().SecretsEnabled(gomock.Any()).Return(true)

	redirect := "https://opsy.test/v1/chat/slack/oauth/callback"
	h.courier.EXPECT().
		AuthorizeURL(gomock.Any(), entity.ChatProviderSlack, entity.SlackOAuthScopes, redirect, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ entity.ChatProvider, _ []string, _ string, state string) (string, error) {
			return "https://slack.com/oauth/v2/authorize?state=" + state, nil
		})

	var stored entity.ChatOAuthState
	var storedKey string
	h.oauthStates.EXPECT().
		Store(gomock.Any(), gomock.Any(), gomock.Any(), entity.ChatOAuthStateTTL).
		DoAndReturn(func(_ context.Context, state string, data entity.ChatOAuthState, _ time.Duration) error {
			storedKey = state
			stored = data
			return nil
		})

	url, err := h.srv.StartOAuth(sessionCtx(), "acme", entity.ChatProviderSlack)
	if err != nil {
		t.Fatalf("StartOAuth: %v", err)
	}
	if storedKey == "" || !strings.Contains(url, storedKey) {
		t.Errorf("authorize URL %q must carry the stored state %q", url, storedKey)
	}
	if stored.WorkspaceID != "ws-1" || stored.WorkspaceSlug != "acme" || stored.UserID != "u1" || stored.Provider != entity.ChatProviderSlack {
		t.Errorf("stored state = %+v, want ws-1/acme/u1/slack", stored)
	}
}

func TestStartOAuthRejectsNonOAuthProvider(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.allowWrite()

	_, err := h.srv.StartOAuth(sessionCtx(), "acme", entity.ChatProviderDiscord)
	if err != entity.ErrChatOAuthUnsupported {
		t.Fatalf("err = %v, want ErrChatOAuthUnsupported for Discord", err)
	}
}

func TestStartOAuthRejectsWhenSecretStorageDisabled(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.allowWrite()
	h.connections.EXPECT().SecretsEnabled(gomock.Any()).Return(false)

	_, err := h.srv.StartOAuth(sessionCtx(), "acme", entity.ChatProviderSlack)
	if err != entity.ErrChatSecretUnavailable {
		t.Fatalf("err = %v, want ErrChatSecretUnavailable so we never send the user to Slack", err)
	}
}

func TestCompleteOAuthExchangesAndSaves(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})

	h.oauthStates.EXPECT().Consume(gomock.Any(), "state-xyz").Return(entity.ChatOAuthState{
		Provider: entity.ChatProviderSlack, WorkspaceID: "ws-1", WorkspaceSlug: "acme", UserID: "u1",
	}, nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", entity.PolicyObjectChat, entity.PolicyActionWrite).Return(true, nil)
	h.courier.EXPECT().
		ExchangeOAuth(gomock.Any(), entity.ChatProviderSlack, "code-abc", "https://opsy.test/v1/chat/slack/oauth/callback").
		Return(entity.ChatOAuthResult{ExternalID: "T9", ExternalName: "Acme", BotUserID: "U0", BotToken: "xoxb-real", Scopes: []string{"chat:write"}}, nil)

	var saved entity.ChatConnectionInput
	h.connections.EXPECT().
		Save(gomock.Any(), "ws-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, in entity.ChatConnectionInput) (entity.ChatConnection, error) {
			saved = in
			return entity.ChatConnection{Provider: in.Provider}, nil
		})
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	slug, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderSlack, "code-abc", "state-xyz")
	if err != nil {
		t.Fatalf("CompleteOAuth: %v", err)
	}
	if slug != "acme" {
		t.Errorf("slug = %q, want acme", slug)
	}
	if saved.BotToken != "xoxb-real" || saved.ExternalID != "T9" || saved.ConnectedBy != "u1" {
		t.Errorf("saved OAuth connection wrong: %+v", saved)
	}
}

func TestCompleteOAuthRejectsInvalidState(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})

	h.oauthStates.EXPECT().Consume(gomock.Any(), "bad").Return(entity.ChatOAuthState{}, entity.ErrChatOAuthStateInvalid)

	slug, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderSlack, "code", "bad")
	if err != entity.ErrChatOAuthStateInvalid {
		t.Fatalf("err = %v, want ErrChatOAuthStateInvalid", err)
	}
	if slug != "" {
		t.Errorf("slug = %q, want empty on invalid state", slug)
	}
}

func TestCompleteOAuthRejectsSessionMismatch(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})

	h.oauthStates.EXPECT().Consume(gomock.Any(), "state-xyz").Return(entity.ChatOAuthState{
		Provider: entity.ChatProviderSlack, WorkspaceID: "ws-1", WorkspaceSlug: "acme", UserID: "u1",
	}, nil)

	ctx := entity.WithIdentity(context.Background(), entity.Identity{Kind: entity.IdentityKindSession, UserID: "attacker"})
	slug, err := h.srv.CompleteOAuth(ctx, entity.ChatProviderSlack, "code-abc", "state-xyz")
	if err != entity.ErrChatOAuthStateInvalid {
		t.Fatalf("err = %v, want ErrChatOAuthStateInvalid when the completing session is not the initiator", err)
	}
	if slug != "acme" {
		t.Errorf("slug = %q, want acme (so the user gets feedback)", slug)
	}
}

func identityState() entity.ChatOAuthState {
	return entity.ChatOAuthState{
		Provider: entity.ChatProviderSlack, Purpose: entity.ChatOAuthIdentity,
		WorkspaceID: "ws-1", WorkspaceSlug: "acme", UserID: "u1", ConnectionID: "conn-1", TeamID: "T9",
	}
}

func TestStartIdentityOAuthStoresIdentityState(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.allowRead()
	h.connections.EXPECT().Get(gomock.Any(), "ws-1", entity.ChatProviderSlack).
		Return(entity.ChatConnection{ID: "conn-1", ExternalID: "T9"}, nil)
	redirect := "https://opsy.test/v1/chat/slack/identity/callback"
	h.courier.EXPECT().
		IdentityAuthorizeURL(gomock.Any(), entity.ChatProviderSlack, entity.SlackOIDCScopes, redirect, gomock.Any(), "T9").
		DoAndReturn(func(_ context.Context, _ entity.ChatProvider, _ []string, _, state, _ string) (string, error) {
			return "https://slack.com/openid/connect/authorize?state=" + state, nil
		})
	var stored entity.ChatOAuthState
	h.oauthStates.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), entity.ChatOAuthStateTTL).
		DoAndReturn(func(_ context.Context, _ string, data entity.ChatOAuthState, _ time.Duration) error {
			stored = data
			return nil
		})

	if _, err := h.srv.StartIdentityOAuth(sessionCtx(), "acme", entity.ChatProviderSlack); err != nil {
		t.Fatalf("StartIdentityOAuth: %v", err)
	}
	if stored.Purpose != entity.ChatOAuthIdentity || stored.ConnectionID != "conn-1" || stored.TeamID != "T9" || stored.UserID != "u1" {
		t.Errorf("stored identity state = %+v", stored)
	}
}

func TestCompleteIdentityOAuthStoresVerifiedIdentity(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.oauthStates.EXPECT().Consume(gomock.Any(), "st").Return(identityState(), nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.courier.EXPECT().
		ExchangeIdentity(gomock.Any(), entity.ChatProviderSlack, "code", "https://opsy.test/v1/chat/slack/identity/callback").
		Return(entity.ChatIdentityResult{ProviderUserID: "U777", TeamID: "T9", Handle: "vlad"}, nil)
	var up entity.ChatIdentity
	h.identities.EXPECT().Upsert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, in entity.ChatIdentity) (entity.ChatIdentity, error) {
			up = in
			return in, nil
		})

	slug, err := h.srv.CompleteIdentityOAuth(sessionCtx(), entity.ChatProviderSlack, "code", "st")
	if err != nil {
		t.Fatalf("CompleteIdentityOAuth: %v", err)
	}
	if slug != "acme" {
		t.Errorf("slug = %q, want acme", slug)
	}
	if up.ProviderUserID != "U777" || up.ConnectionID != "conn-1" || up.UserID != "u1" || up.ResolvedBy != "oauth" || !up.Verified {
		t.Errorf("stored identity wrong: %+v", up)
	}
}

func TestCompleteIdentityOAuthRejectsSessionMismatch(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.oauthStates.EXPECT().Consume(gomock.Any(), "st").Return(identityState(), nil)

	ctx := entity.WithIdentity(context.Background(), entity.Identity{Kind: entity.IdentityKindSession, UserID: "attacker"})
	slug, err := h.srv.CompleteIdentityOAuth(ctx, entity.ChatProviderSlack, "code", "st")
	if err != entity.ErrChatOAuthStateInvalid {
		t.Fatalf("err = %v, want ErrChatOAuthStateInvalid when a different session completes the flow", err)
	}
	if slug != "acme" {
		t.Errorf("slug = %q, want acme", slug)
	}
}

func TestCompleteIdentityOAuthRejectsTeamMismatch(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.oauthStates.EXPECT().Consume(gomock.Any(), "st").Return(identityState(), nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.courier.EXPECT().
		ExchangeIdentity(gomock.Any(), entity.ChatProviderSlack, "code", gomock.Any()).
		Return(entity.ChatIdentityResult{ProviderUserID: "U777", TeamID: "OTHER"}, nil)

	_, err := h.srv.CompleteIdentityOAuth(sessionCtx(), entity.ChatProviderSlack, "code", "st")
	if err != entity.ErrChatOAuthStateInvalid {
		t.Fatalf("err = %v, want ErrChatOAuthStateInvalid when the id_token team != connected team", err)
	}
}

func TestCompleteOAuthRejectsIdentityState(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.oauthStates.EXPECT().Consume(gomock.Any(), "st").Return(identityState(), nil)

	_, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderSlack, "code", "st")
	if err != entity.ErrChatOAuthStateInvalid {
		t.Fatalf("err = %v, want ErrChatOAuthStateInvalid: an identity-purpose state must not install a bot", err)
	}
}

func TestListEnrichesWithLinkedIdentity(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.allowRead()
	h.connections.EXPECT().List(gomock.Any(), "ws-1").
		Return([]entity.ChatConnection{{ID: "conn-1", Provider: entity.ChatProviderSlack}}, nil)
	h.identities.EXPECT().GetForUser(gomock.Any(), "conn-1", "u1").
		Return(entity.ChatIdentity{ProviderHandle: "vlad", Verified: true}, nil)

	conns, err := h.srv.List(sessionCtx(), "acme")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if !conns[0].Linked || conns[0].LinkedHandle != "vlad" || !conns[0].LinkedVerified {
		t.Errorf("connection not enriched with linked identity: %+v", conns[0])
	}
}

func TestTestConnectionLinksOnEmailResolve(t *testing.T) {
	h := newHarness(t, config.Slack{}, config.Discord{})
	h.allowRead()
	h.connections.EXPECT().Get(gomock.Any(), "ws-1", entity.ChatProviderSlack).
		Return(entity.ChatConnection{ID: "conn-1", ExternalID: "T9"}, nil)
	h.connections.EXPECT().BotToken(gomock.Any(), "ws-1", entity.ChatProviderSlack).Return("xoxb", nil)
	h.identities.EXPECT().GetForUser(gomock.Any(), "conn-1", "u1").
		Return(entity.ChatIdentity{}, entity.ErrChatNotConnected)
	h.members.EXPECT().Get(gomock.Any(), "ws-1", "u1").
		Return(entity.Member{Email: "vlad@corp.com"}, nil)
	h.courier.EXPECT().LookupUser(gomock.Any(), entity.ChatProviderSlack, "xoxb", "T9", "vlad@corp.com").
		Return(entity.ChatUser{ProviderUserID: "U9", Handle: "vlad"}, nil)
	h.courier.EXPECT().SendDirect(gomock.Any(), entity.ChatProviderSlack, "xoxb", "U9", "", gomock.Any()).
		Return(entity.ChatSendResult{DMChannelID: "D1", Result: entity.NotifyResult{Delivered: true}}, nil)
	var up entity.ChatIdentity
	h.identities.EXPECT().Upsert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, in entity.ChatIdentity) (entity.ChatIdentity, error) {
			up = in
			return in, nil
		})

	res, err := h.srv.TestConnection(sessionCtx(), "acme", entity.ChatProviderSlack)
	if err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !res.Result.Delivered {
		t.Fatalf("expected delivered")
	}
	if up.ProviderUserID != "U9" || up.ConnectionID != "conn-1" || up.ResolvedBy != "email" || !up.Verified || up.DMChannelID != "D1" {
		t.Errorf("email-resolved test did not link the identity correctly: %+v", up)
	}
}
