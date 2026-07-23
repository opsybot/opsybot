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

func newHarness(t *testing.T, discord config.Discord) *harness {
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
		cfg: config.Auth{BaseURL: "https://opsy.test"}, discord: discord,
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

func TestCompleteOAuthSweepsSlackIdentitiesByEmail(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.oauthStates.EXPECT().Consume(gomock.Any(), "state-xyz").Return(entity.ChatOAuthState{
		Provider: entity.ChatProviderSlack, WorkspaceID: "ws-1", WorkspaceSlug: "acme", UserID: "u1",
	}, nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", entity.PolicyObjectChat, entity.PolicyActionWrite).Return(true, nil)
	h.courier.EXPECT().
		ExchangeOAuth(gomock.Any(), entity.ChatProviderSlack, "code-abc", gomock.Any()).
		Return(entity.ChatOAuthResult{ExternalID: "T9", ExternalName: "Acme", BotUserID: "U0", BotToken: "xoxb-real"}, nil)
	h.connections.EXPECT().Save(gomock.Any(), "ws-1", gomock.Any()).
		Return(entity.ChatConnection{ID: "conn-1", WorkspaceID: "ws-1", Provider: entity.ChatProviderSlack, ExternalID: "T9"}, nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	h.members.EXPECT().ListByWorkspace(gomock.Any(), "ws-1").Return([]entity.Member{
		{UserID: "u2", Email: "priya@acme.test", Status: entity.MemberStatusActive},
		{UserID: "u3", Email: "", Status: entity.MemberStatusActive},
		{UserID: "u4", Email: "old@acme.test", Status: entity.MemberStatusInvited},
	}, nil)
	h.courier.EXPECT().
		LookupUser(gomock.Any(), entity.ChatProviderSlack, "xoxb-real", "T9", "priya@acme.test").
		Return(entity.ChatUser{ProviderUserID: "U100", Handle: "priya"}, nil)
	h.identities.EXPECT().Upsert(gomock.Any(), gomock.Cond(func(in entity.ChatIdentity) bool {
		return in.ConnectionID == "conn-1" && in.UserID == "u2" && in.ProviderUserID == "U100" &&
			in.ResolvedBy == "email" && in.Verified
	})).Return(entity.ChatIdentity{}, nil)

	if _, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderSlack, "code-abc", "", "state-xyz"); err != nil {
		t.Fatalf("CompleteOAuth: %v", err)
	}
}

func TestConnectTeamsValidatesWithoutTokenAndSweepsByEmail(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowWrite()
	h.courier.EXPECT().
		Validate(gomock.Any(), entity.ChatProviderTeams, "", "").
		Return(entity.ChatValidation{ExternalID: "tenant-1", ExternalName: "Microsoft Teams"}, nil)
	h.connections.EXPECT().Save(gomock.Any(), "ws-1", gomock.Any()).
		Return(entity.ChatConnection{ID: "conn-t", WorkspaceID: "ws-1", Provider: entity.ChatProviderTeams, ExternalID: "tenant-1"}, nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	h.members.EXPECT().ListByWorkspace(gomock.Any(), "ws-1").Return([]entity.Member{
		{UserID: "u1", Email: "vlad@acme.test", Status: entity.MemberStatusActive},
	}, nil)
	h.courier.EXPECT().
		LookupUser(gomock.Any(), entity.ChatProviderTeams, "", "tenant-1", "vlad@acme.test").
		Return(entity.ChatUser{ProviderUserID: "aad-1", Handle: "Vlad"}, nil)
	h.identities.EXPECT().Upsert(gomock.Any(), gomock.Cond(func(in entity.ChatIdentity) bool {
		return in.ConnectionID == "conn-t" && in.UserID == "u1" && in.ProviderUserID == "aad-1" &&
			in.ResolvedBy == "email" && in.Verified
	})).Return(entity.ChatIdentity{}, nil)

	if _, err := h.srv.Connect(sessionCtx(), "acme", entity.ChatConnectInput{Provider: entity.ChatProviderTeams}); err != nil {
		t.Fatalf("Connect(teams): %v", err)
	}
}

func TestConnectRejectsWhenNoTokenAvailable(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowWrite()

	_, err := h.srv.Connect(sessionCtx(), "acme", entity.ChatConnectInput{Provider: entity.ChatProviderSlack})
	if err != entity.ErrChatProviderNotConfigured {
		t.Fatalf("err = %v, want ErrChatProviderNotConfigured when neither UI nor env supplies a token", err)
	}
}

func TestStartOAuthStoresStateAndReturnsURL(t *testing.T) {
	h := newHarness(t, config.Discord{})
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
	h := newHarness(t, config.Discord{})
	h.allowWrite()

	_, err := h.srv.StartOAuth(sessionCtx(), "acme", entity.ChatProviderTeams)
	if err != entity.ErrChatOAuthUnsupported {
		t.Fatalf("err = %v, want ErrChatOAuthUnsupported for Teams", err)
	}
}

func TestStartOAuthDiscordUsesBotInviteNoSecretCheck(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowWrite()
	// Discord stores no secret, so it must NOT gate on secret storage.
	redirect := "https://opsy.test/v1/chat/discord/oauth/callback"
	h.courier.EXPECT().
		AuthorizeURL(gomock.Any(), entity.ChatProviderDiscord, entity.DiscordBotScopes, redirect, gomock.Any()).
		Return("https://discord.com/oauth2/authorize?scope=bot", nil)
	h.oauthStates.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), entity.ChatOAuthStateTTL).Return(nil)

	if _, err := h.srv.StartOAuth(sessionCtx(), "acme", entity.ChatProviderDiscord); err != nil {
		t.Fatalf("StartOAuth(discord): %v", err)
	}
}

func TestCompleteOAuthDiscordStoresGuildNoToken(t *testing.T) {
	h := newHarness(t, config.Discord{BotToken: "disc-env"})
	h.oauthStates.EXPECT().Consume(gomock.Any(), "st").Return(entity.ChatOAuthState{
		Provider: entity.ChatProviderDiscord, Purpose: entity.ChatOAuthInstall,
		WorkspaceID: "ws-1", WorkspaceSlug: "acme", UserID: "u1",
	}, nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", entity.PolicyObjectChat, entity.PolicyActionWrite).Return(true, nil)
	h.courier.EXPECT().
		Validate(gomock.Any(), entity.ChatProviderDiscord, "disc-env", "G42").
		Return(entity.ChatValidation{ExternalID: "G42", ExternalName: "Vlad's Server", BotUserID: "Bot0"}, nil)
	var saved entity.ChatConnectionInput
	h.connections.EXPECT().Save(gomock.Any(), "ws-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, in entity.ChatConnectionInput) (entity.ChatConnection, error) {
			saved = in
			return entity.ChatConnection{Provider: in.Provider}, nil
		})
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	slug, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderDiscord, "code", "G42", "st")
	if err != nil {
		t.Fatalf("CompleteOAuth(discord): %v", err)
	}
	if slug != "acme" {
		t.Errorf("slug = %q, want acme", slug)
	}
	if saved.ExternalID != "G42" || saved.ExternalName != "Vlad's Server" || saved.BotToken != "" {
		t.Errorf("discord install saved wrong (guild id from callback, no token stored): %+v", saved)
	}
}

func TestCompleteOAuthDiscordRequiresGuild(t *testing.T) {
	h := newHarness(t, config.Discord{BotToken: "disc-env"})
	h.oauthStates.EXPECT().Consume(gomock.Any(), "st").Return(entity.ChatOAuthState{
		Provider: entity.ChatProviderDiscord, Purpose: entity.ChatOAuthInstall,
		WorkspaceID: "ws-1", WorkspaceSlug: "acme", UserID: "u1",
	}, nil)
	h.members.EXPECT().IsActive(gomock.Any(), "ws-1", "u1").Return(true, nil)
	h.pol.EXPECT().Allowed(gomock.Any(), gomock.Any(), "ws-1", entity.PolicyObjectChat, entity.PolicyActionWrite).Return(true, nil)

	_, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderDiscord, "code", "", "st")
	if err != entity.ErrChatOAuthExchange {
		t.Fatalf("err = %v, want ErrChatOAuthExchange when Discord returns no guild_id", err)
	}
}

func TestStartOAuthRejectsWhenSecretStorageDisabled(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowWrite()
	h.connections.EXPECT().SecretsEnabled(gomock.Any()).Return(false)

	_, err := h.srv.StartOAuth(sessionCtx(), "acme", entity.ChatProviderSlack)
	if err != entity.ErrChatSecretUnavailable {
		t.Fatalf("err = %v, want ErrChatSecretUnavailable so we never send the user to Slack", err)
	}
}

func TestCompleteOAuthExchangesAndSaves(t *testing.T) {
	h := newHarness(t, config.Discord{})

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
	h.members.EXPECT().ListByWorkspace(gomock.Any(), gomock.Any()).Return(nil, nil)

	slug, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderSlack, "code-abc", "", "state-xyz")
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
	h := newHarness(t, config.Discord{})

	h.oauthStates.EXPECT().Consume(gomock.Any(), "bad").Return(entity.ChatOAuthState{}, entity.ErrChatOAuthStateInvalid)

	slug, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderSlack, "code", "", "bad")
	if err != entity.ErrChatOAuthStateInvalid {
		t.Fatalf("err = %v, want ErrChatOAuthStateInvalid", err)
	}
	if slug != "" {
		t.Errorf("slug = %q, want empty on invalid state", slug)
	}
}

func TestCompleteOAuthRejectsSessionMismatch(t *testing.T) {
	h := newHarness(t, config.Discord{})

	h.oauthStates.EXPECT().Consume(gomock.Any(), "state-xyz").Return(entity.ChatOAuthState{
		Provider: entity.ChatProviderSlack, WorkspaceID: "ws-1", WorkspaceSlug: "acme", UserID: "u1",
	}, nil)

	ctx := entity.WithIdentity(context.Background(), entity.Identity{Kind: entity.IdentityKindSession, UserID: "attacker"})
	slug, err := h.srv.CompleteOAuth(ctx, entity.ChatProviderSlack, "code-abc", "", "state-xyz")
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
	h := newHarness(t, config.Discord{})
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
	h := newHarness(t, config.Discord{})
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
	h := newHarness(t, config.Discord{})
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
	h := newHarness(t, config.Discord{})
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
	h := newHarness(t, config.Discord{})
	h.oauthStates.EXPECT().Consume(gomock.Any(), "st").Return(identityState(), nil)

	_, err := h.srv.CompleteOAuth(sessionCtx(), entity.ChatProviderSlack, "code", "", "st")
	if err != entity.ErrChatOAuthStateInvalid {
		t.Fatalf("err = %v, want ErrChatOAuthStateInvalid: an identity-purpose state must not install a bot", err)
	}
}

func TestListEnrichesWithLinkedIdentity(t *testing.T) {
	h := newHarness(t, config.Discord{})
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

func TestTestConnectionPostsToAnnounceChannel(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowRead()
	h.connections.EXPECT().Get(gomock.Any(), "ws-1", entity.ChatProviderSlack).
		Return(entity.ChatConnection{ID: "conn-1", ExternalID: "T9", AnnounceChannel: "#ops"}, nil)
	h.connections.EXPECT().BotToken(gomock.Any(), "ws-1", entity.ChatProviderSlack).Return("xoxb", nil)
	h.courier.EXPECT().
		SendToChannel(gomock.Any(), entity.ChatProviderSlack, "xoxb", "T9", "#ops", gomock.Any()).
		Return(entity.ChatSendResult{Result: entity.NotifyResult{Delivered: true}}, nil)

	res, err := h.srv.TestConnection(sessionCtx(), "acme", entity.ChatProviderSlack)
	if err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !res.Result.Delivered {
		t.Fatalf("expected delivered")
	}
	if res.Result.Detail != "Posted a test message to #ops." {
		t.Errorf("detail = %q, want channel confirmation", res.Result.Detail)
	}
}

func TestTestConnectionFallsBackToDefaultChannel(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowRead()
	// No announce channel configured -> uses the default.
	h.connections.EXPECT().Get(gomock.Any(), "ws-1", entity.ChatProviderDiscord).
		Return(entity.ChatConnection{ID: "conn-1", ExternalID: "G9"}, nil)
	h.connections.EXPECT().BotToken(gomock.Any(), "ws-1", entity.ChatProviderDiscord).Return("bot", nil)
	h.courier.EXPECT().
		SendToChannel(gomock.Any(), entity.ChatProviderDiscord, "bot", "G9", entity.DefaultAnnounceChannel, gomock.Any()).
		Return(entity.ChatSendResult{Result: entity.NotifyResult{Delivered: true}}, nil)

	if _, err := h.srv.TestConnection(sessionCtx(), "acme", entity.ChatProviderDiscord); err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
}

func TestStartTelegramLinkReturnsDeepLink(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowRead()
	h.connections.EXPECT().Get(gomock.Any(), "ws-1", entity.ChatProviderTelegram).
		Return(entity.ChatConnection{ID: "conn-1", ExternalName: "opsy_bot"}, nil)
	var storedKey string
	var stored entity.ChatOAuthState
	h.oauthStates.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), entity.ChatOAuthStateTTL).
		DoAndReturn(func(_ context.Context, key string, data entity.ChatOAuthState, _ time.Duration) error {
			storedKey, stored = key, data
			return nil
		})

	url, err := h.srv.StartTelegramLink(sessionCtx(), "acme")
	if err != nil {
		t.Fatalf("StartTelegramLink: %v", err)
	}
	if url != "https://t.me/opsy_bot?start="+storedKey {
		t.Errorf("deep link = %q", url)
	}
	if stored.Purpose != entity.ChatOAuthLink || stored.Provider != entity.ChatProviderTelegram || stored.ConnectionID != "conn-1" || stored.UserID != "u1" {
		t.Errorf("stored link state = %+v", stored)
	}
}

func TestCompleteTelegramLinkStoresIdentity(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.srv.telegram = config.Telegram{BotToken: "tg-token"}
	h.oauthStates.EXPECT().Consume(gomock.Any(), "linktok").Return(entity.ChatOAuthState{
		Provider: entity.ChatProviderTelegram, Purpose: entity.ChatOAuthLink,
		WorkspaceID: "ws-1", WorkspaceSlug: "acme", UserID: "u1", ConnectionID: "conn-1",
	}, nil)
	var up entity.ChatIdentity
	h.identities.EXPECT().Upsert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, in entity.ChatIdentity) (entity.ChatIdentity, error) {
			up = in
			return in, nil
		})
	h.courier.EXPECT().
		SendToChannel(gomock.Any(), entity.ChatProviderTelegram, "tg-token", "", "88888", gomock.Any()).
		Return(entity.ChatSendResult{Result: entity.NotifyResult{Delivered: true}}, nil)

	if err := h.srv.CompleteTelegramLink(context.Background(), "linktok", "88888", "vlad_tg"); err != nil {
		t.Fatalf("CompleteTelegramLink: %v", err)
	}
	if up.ProviderUserID != "88888" || up.DMChannelID != "88888" || up.ConnectionID != "conn-1" || up.ResolvedBy != "telegram" || !up.Verified {
		t.Errorf("telegram identity stored wrong: %+v", up)
	}
}

func TestConnectTelegramSetsWebhook(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.srv.telegram = config.Telegram{BotToken: "tg-token"}
	h.allowWrite()
	h.courier.EXPECT().Validate(gomock.Any(), entity.ChatProviderTelegram, "tg-token", "").
		Return(entity.ChatValidation{ExternalID: "42", ExternalName: "opsy_bot", BotUserID: "42"}, nil)
	h.courier.EXPECT().SetWebhook(gomock.Any(), entity.ChatProviderTelegram, "tg-token", gomock.Any(), gomock.Any()).Return(nil)
	var saved entity.ChatConnectionInput
	h.connections.EXPECT().Save(gomock.Any(), "ws-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, in entity.ChatConnectionInput) (entity.ChatConnection, error) {
			saved = in
			return entity.ChatConnection{Provider: in.Provider}, nil
		})
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	if _, err := h.srv.Connect(sessionCtx(), "acme", entity.ChatConnectInput{Provider: entity.ChatProviderTelegram}); err != nil {
		t.Fatalf("Connect(telegram): %v", err)
	}
	if saved.ExternalName != "opsy_bot" || saved.BotToken != "" {
		t.Errorf("telegram connection saved wrong (env token, not stored): %+v", saved)
	}
}

func TestTestConnectionTelegramDMsTheUser(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowRead()
	h.connections.EXPECT().Get(gomock.Any(), "ws-1", entity.ChatProviderTelegram).
		Return(entity.ChatConnection{ID: "conn-1", ExternalName: "opsy_bot"}, nil)
	h.connections.EXPECT().BotToken(gomock.Any(), "ws-1", entity.ChatProviderTelegram).Return("tg", nil)
	h.identities.EXPECT().GetForUser(gomock.Any(), "conn-1", "u1").
		Return(entity.ChatIdentity{ProviderUserID: "426937641"}, nil)
	h.courier.EXPECT().
		SendToChannel(gomock.Any(), entity.ChatProviderTelegram, "tg", "", "426937641", gomock.Any()).
		Return(entity.ChatSendResult{Result: entity.NotifyResult{Delivered: true}}, nil)

	res, err := h.srv.TestConnection(sessionCtx(), "acme", entity.ChatProviderTelegram)
	if err != nil {
		t.Fatalf("TestConnection(telegram): %v", err)
	}
	if !res.Result.Delivered || res.Result.Detail != "Sent you a test message on Telegram." {
		t.Errorf("telegram test should DM the user: %+v", res.Result)
	}
}

func TestTestConnectionTelegramNeedsLinkFirst(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.allowRead()
	h.connections.EXPECT().Get(gomock.Any(), "ws-1", entity.ChatProviderTelegram).
		Return(entity.ChatConnection{ID: "conn-1"}, nil)
	h.connections.EXPECT().BotToken(gomock.Any(), "ws-1", entity.ChatProviderTelegram).Return("tg", nil)
	h.identities.EXPECT().GetForUser(gomock.Any(), "conn-1", "u1").
		Return(entity.ChatIdentity{}, entity.ErrChatNotConnected)

	res, err := h.srv.TestConnection(sessionCtx(), "acme", entity.ChatProviderTelegram)
	if err != nil {
		t.Fatalf("TestConnection(telegram): %v", err)
	}
	if res.Result.Delivered || res.Result.Detail != "Link your Telegram account first, then test." {
		t.Errorf("unlinked telegram test should ask to link first: %+v", res.Result)
	}
}

func TestAnswerTelegramCallbackDelegates(t *testing.T) {
	h := newHarness(t, config.Discord{})
	h.srv.telegram = config.Telegram{BotToken: "tg"}
	h.courier.EXPECT().
		AnswerCallback(gomock.Any(), entity.ChatProviderTelegram, "tg", "cbid", "Acknowledged: DB down").
		Return(nil)

	if err := h.srv.AnswerTelegramCallback(context.Background(), "cbid", "Acknowledged: DB down"); err != nil {
		t.Fatalf("AnswerTelegramCallback: %v", err)
	}
}
