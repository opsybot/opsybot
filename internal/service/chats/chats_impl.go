package chats

import (
	"context"
	"strings"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx          repository.Transactor
	workspaces  repository.Workspace
	members     repository.Member
	policy      repository.Policy
	connections repository.ChatConnection
	identities  repository.ChatIdentity
	courier     repository.ChatCourier
	oauthStates repository.ChatOAuthState
	audit       repository.Audit
	cfg         config.Auth
	slack       config.Slack
	discord     config.Discord
	telegram    config.Telegram
}

func New(
	tx repository.Transactor,
	workspaces repository.Workspace,
	members repository.Member,
	policy repository.Policy,
	connections repository.ChatConnection,
	identities repository.ChatIdentity,
	courier repository.ChatCourier,
	oauthStates repository.ChatOAuthState,
	audit repository.Audit,
	cfg config.Auth,
	slack config.Slack,
	discord config.Discord,
	telegram config.Telegram,
) service.Chats {
	return &srv{tx: tx, workspaces: workspaces, members: members, policy: policy, connections: connections, identities: identities, courier: courier, oauthStates: oauthStates, audit: audit, cfg: cfg, slack: slack, discord: discord, telegram: telegram}
}

func (s *srv) redirectURI(provider entity.ChatProvider) string {
	return strings.TrimRight(s.cfg.BaseURL, "/") + "/v1/chat/" + string(provider) + "/oauth/callback"
}

func (s *srv) envToken(provider entity.ChatProvider) string {
	switch provider {
	case entity.ChatProviderSlack:
		return s.slack.BotToken
	case entity.ChatProviderDiscord:
		return s.discord.BotToken
	case entity.ChatProviderTelegram:
		return s.telegram.BotToken
	default:
		return ""
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
	if !id.ScopePermits(entity.PolicyObjectChat, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectChat, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string) ([]entity.ChatConnection, error) {
	id, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return nil, err
	}
	conns, err := s.connections.List(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	if id.Kind == entity.IdentityKindSession && id.UserID != "" {
		for i := range conns {
			ident, iErr := s.identities.GetForUser(ctx, conns[i].ID, id.UserID)
			if iErr != nil {
				continue
			}
			conns[i].Linked = true
			conns[i].LinkedHandle = ident.ProviderHandle
			conns[i].LinkedVerified = ident.Verified
			conns[i].LinkedMethod = ident.ResolvedBy
		}
	}
	return conns, nil
}

func (s *srv) Connect(ctx context.Context, workspaceSlug string, in entity.ChatConnectInput) (entity.ChatConnection, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.ChatConnection{}, err
	}
	if !in.Provider.Valid() {
		return entity.ChatConnection{}, entity.ErrChatConnectionInvalid
	}
	token := in.BotToken
	storedToken := in.BotToken
	if token == "" {
		token = s.envToken(in.Provider)
		storedToken = ""
	}
	if token == "" {
		return entity.ChatConnection{}, entity.ErrChatProviderNotConfigured
	}
	valid, err := s.courier.Validate(ctx, in.Provider, token, in.ExternalID)
	if err != nil {
		return entity.ChatConnection{}, err
	}
	if in.Provider == entity.ChatProviderTelegram {
		secret := entity.TelegramWebhookSecret(token)
		hookURL := strings.TrimRight(s.cfg.BaseURL, "/") + "/v1/chat/telegram/hook/" + secret
		if err := s.courier.SetWebhook(ctx, in.Provider, token, hookURL, secret); err != nil {
			return entity.ChatConnection{}, entity.ErrChatConnectionInvalid
		}
	}
	conn, err := s.connections.Save(ctx, ws.ID, entity.ChatConnectionInput{
		Provider: in.Provider, ExternalID: valid.ExternalID, ExternalName: valid.ExternalName,
		BotUserID: valid.BotUserID, BotToken: storedToken, ConnectedBy: actor.UserID,
		ArchiveOnResolve: true,
	})
	if err != nil {
		return entity.ChatConnection{}, err
	}
	_ = s.audit.Create(ctx, entity.AuditEvent{
		WorkspaceID: ws.ID, ActorType: entity.AuditActorUser, ActorUserID: actor.UserID,
		ActorLabel: actor.Label, Action: entity.ActionChatConnected, Target: string(in.Provider),
	})
	if in.Provider == entity.ChatProviderSlack {
		s.sweepSlackIdentities(ctx, conn, token)
	}
	return conn, nil
}

func (s *srv) sweepSlackIdentities(ctx context.Context, conn entity.ChatConnection, token string) {
	members, err := s.members.ListByWorkspace(ctx, conn.WorkspaceID)
	if err != nil {
		logger.From(ctx).WarnContext(ctx, "slack identity sweep skipped", "error", err, "workspace_id", conn.WorkspaceID)
		return
	}
	for _, m := range members {
		if m.Status != entity.MemberStatusActive || m.Email == "" {
			continue
		}
		user, err := s.courier.LookupUser(ctx, entity.ChatProviderSlack, token, conn.ExternalID, m.Email)
		if err != nil || user.ProviderUserID == "" {
			continue
		}
		if _, err := s.identities.Upsert(ctx, entity.ChatIdentity{
			ConnectionID: conn.ID, UserID: m.UserID, ProviderUserID: user.ProviderUserID,
			ProviderHandle: user.Handle, ResolvedBy: "email", Verified: true,
		}); err != nil {
			logger.From(ctx).WarnContext(ctx, "slack identity upsert failed", "error", err, "user_id", m.UserID)
		}
	}
}

func (s *srv) Delete(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	if err := s.connections.Delete(ctx, ws.ID, provider); err != nil {
		return err
	}
	_ = s.audit.Create(ctx, entity.AuditEvent{
		WorkspaceID: ws.ID, ActorType: entity.AuditActorUser, ActorUserID: actor.UserID,
		ActorLabel: actor.Label, Action: entity.ActionChatDisconnected, Target: string(provider),
	})
	return nil
}

func (s *srv) SetDefaults(ctx context.Context, workspaceSlug string, provider entity.ChatProvider, namingPattern, announceChannel string, archiveOnResolve bool) error {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	return s.connections.SetDefaults(ctx, ws.ID, provider, namingPattern, announceChannel, archiveOnResolve)
}

func (s *srv) LinkIdentity(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) (entity.ChatIdentity, error) {
	id, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.ChatIdentity{}, err
	}
	if id.Kind != entity.IdentityKindSession {
		return entity.ChatIdentity{}, entity.ErrUnauthenticated
	}
	conn, err := s.connections.Get(ctx, ws.ID, provider)
	if err != nil {
		return entity.ChatIdentity{}, err
	}
	token, err := s.connections.BotToken(ctx, ws.ID, provider)
	if err != nil || token == "" {
		return entity.ChatIdentity{}, entity.ErrChatNotConnected
	}
	member, err := s.members.Get(ctx, ws.ID, id.UserID)
	if err != nil {
		return entity.ChatIdentity{}, err
	}
	user, err := s.courier.LookupUser(ctx, provider, token, conn.ExternalID, member.Email)
	if err != nil {
		return entity.ChatIdentity{}, err
	}
	return s.identities.Upsert(ctx, entity.ChatIdentity{
		ConnectionID: conn.ID, UserID: id.UserID, ProviderUserID: user.ProviderUserID,
		ProviderHandle: user.Handle, ResolvedBy: "email", Verified: true,
	})
}

func (s *srv) TestConnection(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) (entity.ChatSendResult, error) {
	id, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.ChatSendResult{}, err
	}
	if id.Kind != entity.IdentityKindSession {
		return entity.ChatSendResult{}, entity.ErrUnauthenticated
	}
	conn, err := s.connections.Get(ctx, ws.ID, provider)
	if err != nil {
		return entity.ChatSendResult{}, err
	}
	token, err := s.connections.BotToken(ctx, ws.ID, provider)
	if err != nil || token == "" {
		return entity.ChatSendResult{}, entity.ErrChatNotConnected
	}
	if provider == entity.ChatProviderTelegram {
		ident, iErr := s.identities.GetForUser(ctx, conn.ID, id.UserID)
		if iErr != nil {
			return entity.ChatSendResult{Result: entity.NotifyResult{Detail: "Link your Telegram account first, then test."}}, nil
		}
		result, err := s.courier.SendToChannel(ctx, provider, token, "", ident.ProviderUserID,
			"Opsybot test — your Telegram alerts are working. Pages will arrive here.")
		if err != nil {
			return entity.ChatSendResult{}, err
		}
		if result.Result.Delivered {
			result.Result.Detail = "Sent you a test message on Telegram."
		}
		return result, nil
	}
	channel := conn.AnnounceChannel
	if channel == "" {
		channel = entity.DefaultAnnounceChannel
	}
	text := "Opsybot test — " + string(provider) + " is connected to " + ws.Name + ". Alerts and announcements will post to this channel."
	result, err := s.courier.SendToChannel(ctx, provider, token, conn.ExternalID, channel, text)
	if err != nil {
		return entity.ChatSendResult{}, err
	}
	if result.Result.Delivered {
		result.Result.Detail = "Posted a test message to " + channel + "."
	}
	return result, nil
}

func (s *srv) StartOAuth(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) (string, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return "", err
	}
	var scopes []string
	switch provider {
	case entity.ChatProviderSlack:
		scopes = entity.SlackOAuthScopes
	case entity.ChatProviderDiscord:
		scopes = entity.DiscordBotScopes
	default:
		return "", entity.ErrChatOAuthUnsupported
	}
	if provider == entity.ChatProviderSlack && !s.connections.SecretsEnabled(ctx) {
		return "", entity.ErrChatSecretUnavailable
	}
	state, err := entity.GenerateToken(entity.ChatOAuthStateLength)
	if err != nil {
		return "", err
	}
	url, err := s.courier.AuthorizeURL(ctx, provider, scopes, s.redirectURI(provider), state)
	if err != nil {
		return "", err
	}
	if err := s.oauthStates.Store(ctx, state, entity.ChatOAuthState{
		Provider: provider, Purpose: entity.ChatOAuthInstall, WorkspaceID: ws.ID, WorkspaceSlug: ws.Slug, UserID: actor.UserID,
	}, entity.ChatOAuthStateTTL); err != nil {
		return "", err
	}
	return url, nil
}

func (s *srv) CompleteOAuth(ctx context.Context, provider entity.ChatProvider, code, guildID, state string) (string, error) {
	st, err := s.oauthStates.Consume(ctx, state)
	if err != nil {
		return "", err
	}
	if st.Provider != provider || st.Purpose == entity.ChatOAuthIdentity {
		return st.WorkspaceSlug, entity.ErrChatOAuthStateInvalid
	}
	id, ok := entity.IdentityFrom(ctx)
	if !ok || id.UserID == "" || id.UserID != st.UserID {
		return st.WorkspaceSlug, entity.ErrChatOAuthStateInvalid
	}
	active, err := s.members.IsActive(ctx, st.WorkspaceID, st.UserID)
	if err != nil {
		return st.WorkspaceSlug, err
	}
	if !active {
		return st.WorkspaceSlug, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), st.WorkspaceID, entity.PolicyObjectChat, entity.PolicyActionWrite)
	if err != nil {
		return st.WorkspaceSlug, err
	}
	if !allowed {
		return st.WorkspaceSlug, entity.ErrForbidden
	}
	in := entity.ChatConnectionInput{Provider: provider, ConnectedBy: st.UserID, ArchiveOnResolve: true}
	switch provider {
	case entity.ChatProviderSlack:
		result, xErr := s.courier.ExchangeOAuth(ctx, provider, code, s.redirectURI(provider))
		if xErr != nil {
			return st.WorkspaceSlug, xErr
		}
		in.ExternalID, in.ExternalName = result.ExternalID, result.ExternalName
		in.BotUserID, in.BotToken, in.Scopes = result.BotUserID, result.BotToken, result.Scopes
	case entity.ChatProviderDiscord:
		if guildID == "" {
			return st.WorkspaceSlug, entity.ErrChatOAuthExchange
		}
		token := s.envToken(entity.ChatProviderDiscord)
		if token == "" {
			return st.WorkspaceSlug, entity.ErrChatProviderNotConfigured
		}
		valid, vErr := s.courier.Validate(ctx, provider, token, guildID)
		if vErr != nil {
			return st.WorkspaceSlug, vErr
		}
		in.ExternalID, in.ExternalName, in.BotUserID = valid.ExternalID, valid.ExternalName, valid.BotUserID
	default:
		return st.WorkspaceSlug, entity.ErrChatOAuthUnsupported
	}
	savedConn, err := s.connections.Save(ctx, st.WorkspaceID, in)
	if err != nil {
		return st.WorkspaceSlug, err
	}
	_ = s.audit.Create(ctx, entity.AuditEvent{
		WorkspaceID: st.WorkspaceID, ActorType: entity.AuditActorUser, ActorUserID: st.UserID,
		Action: entity.ActionChatConnected, Target: string(provider),
	})
	if provider == entity.ChatProviderSlack {
		s.sweepSlackIdentities(ctx, savedConn, in.BotToken)
	}
	return st.WorkspaceSlug, nil
}

func (s *srv) identityRedirectURI(provider entity.ChatProvider) string {
	return strings.TrimRight(s.cfg.BaseURL, "/") + "/v1/chat/" + string(provider) + "/identity/callback"
}

func (s *srv) StartIdentityOAuth(ctx context.Context, workspaceSlug string, provider entity.ChatProvider) (string, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return "", err
	}
	if actor.Kind != entity.IdentityKindSession {
		return "", entity.ErrUnauthenticated
	}
	var scopes []string
	switch provider {
	case entity.ChatProviderSlack:
		scopes = entity.SlackOIDCScopes
	case entity.ChatProviderDiscord:
		scopes = entity.DiscordIdentityScopes
	default:
		return "", entity.ErrChatOAuthUnsupported
	}
	conn, err := s.connections.Get(ctx, ws.ID, provider)
	if err != nil {
		return "", err
	}
	state, err := entity.GenerateToken(entity.ChatOAuthStateLength)
	if err != nil {
		return "", err
	}
	url, err := s.courier.IdentityAuthorizeURL(ctx, provider, scopes, s.identityRedirectURI(provider), state, conn.ExternalID)
	if err != nil {
		return "", err
	}
	if err := s.oauthStates.Store(ctx, state, entity.ChatOAuthState{
		Provider: provider, Purpose: entity.ChatOAuthIdentity, WorkspaceID: ws.ID, WorkspaceSlug: ws.Slug,
		UserID: actor.UserID, ConnectionID: conn.ID, TeamID: conn.ExternalID,
	}, entity.ChatOAuthStateTTL); err != nil {
		return "", err
	}
	return url, nil
}

func (s *srv) CompleteIdentityOAuth(ctx context.Context, provider entity.ChatProvider, code, state string) (string, error) {
	st, err := s.oauthStates.Consume(ctx, state)
	if err != nil {
		return "", err
	}
	if st.Provider != provider || st.Purpose != entity.ChatOAuthIdentity {
		return st.WorkspaceSlug, entity.ErrChatOAuthStateInvalid
	}
	id, ok := entity.IdentityFrom(ctx)
	if !ok || id.UserID == "" || id.UserID != st.UserID {
		return st.WorkspaceSlug, entity.ErrChatOAuthStateInvalid
	}
	active, err := s.members.IsActive(ctx, st.WorkspaceID, st.UserID)
	if err != nil {
		return st.WorkspaceSlug, err
	}
	if !active {
		return st.WorkspaceSlug, entity.ErrForbidden
	}
	result, err := s.courier.ExchangeIdentity(ctx, provider, code, s.identityRedirectURI(provider))
	if err != nil {
		return st.WorkspaceSlug, err
	}
	if st.TeamID != "" && result.TeamID != "" && result.TeamID != st.TeamID {
		return st.WorkspaceSlug, entity.ErrChatOAuthStateInvalid
	}
	if _, err := s.identities.Upsert(ctx, entity.ChatIdentity{
		ConnectionID: st.ConnectionID, UserID: st.UserID, ProviderUserID: result.ProviderUserID,
		ProviderHandle: result.Handle, ResolvedBy: "oauth", Verified: true,
	}); err != nil {
		return st.WorkspaceSlug, err
	}
	return st.WorkspaceSlug, nil
}

func (s *srv) StartTelegramLink(ctx context.Context, workspaceSlug string) (string, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return "", err
	}
	if actor.Kind != entity.IdentityKindSession {
		return "", entity.ErrUnauthenticated
	}
	conn, err := s.connections.Get(ctx, ws.ID, entity.ChatProviderTelegram)
	if err != nil {
		return "", err
	}
	botName := conn.ExternalName
	if botName == "" {
		botName = s.telegram.BotName
	}
	if botName == "" {
		return "", entity.ErrChatProviderNotConfigured
	}
	state, err := entity.GenerateToken(entity.ChatOAuthStateLength)
	if err != nil {
		return "", err
	}
	if err := s.oauthStates.Store(ctx, state, entity.ChatOAuthState{
		Provider: entity.ChatProviderTelegram, Purpose: entity.ChatOAuthLink,
		WorkspaceID: ws.ID, WorkspaceSlug: ws.Slug, UserID: actor.UserID, ConnectionID: conn.ID,
	}, entity.ChatOAuthStateTTL); err != nil {
		return "", err
	}
	return "https://t.me/" + botName + "?start=" + state, nil
}

func (s *srv) CompleteTelegramLink(ctx context.Context, token, telegramUserID, handle string) error {
	st, err := s.oauthStates.Consume(ctx, token)
	if err != nil {
		return err
	}
	if st.Provider != entity.ChatProviderTelegram || st.Purpose != entity.ChatOAuthLink {
		return entity.ErrChatOAuthStateInvalid
	}
	if _, err := s.identities.Upsert(ctx, entity.ChatIdentity{
		ConnectionID: st.ConnectionID, UserID: st.UserID, ProviderUserID: telegramUserID,
		ProviderHandle: handle, DMChannelID: telegramUserID, ResolvedBy: "telegram", Verified: true,
	}); err != nil {
		return err
	}
	if botToken := s.envToken(entity.ChatProviderTelegram); botToken != "" {
		_, _ = s.courier.SendToChannel(ctx, entity.ChatProviderTelegram, botToken, "", telegramUserID,
			"Your Telegram is now linked to Opsybot. You'll receive your alerts here.")
	}
	return nil
}

func (s *srv) AnswerTelegramCallback(ctx context.Context, callbackID, text string) error {
	token := s.envToken(entity.ChatProviderTelegram)
	if token == "" {
		return nil
	}
	return s.courier.AnswerCallback(ctx, entity.ChatProviderTelegram, token, callbackID, text)
}
