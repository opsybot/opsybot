package sso

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	cfg         config.Auth
	tx          repository.Transactor
	workspaces  repository.Workspace
	members     repository.Member
	users       repository.User
	policy      repository.Policy
	sessions    repository.Session
	audit       repository.Audit
	connections repository.SSOConnection
	identities  repository.UserIdentity
	states      repository.SSOState

	mu        sync.Mutex
	providers map[string]*oidc.Provider
}

func New(
	cfg config.Auth,
	tx repository.Transactor,
	workspaces repository.Workspace,
	members repository.Member,
	users repository.User,
	policy repository.Policy,
	sessions repository.Session,
	audit repository.Audit,
	connections repository.SSOConnection,
	identities repository.UserIdentity,
	states repository.SSOState,
) service.SSO {
	return &srv{
		cfg: cfg, tx: tx, workspaces: workspaces, members: members, users: users, policy: policy,
		sessions: sessions, audit: audit, connections: connections, identities: identities, states: states,
		providers: make(map[string]*oidc.Provider),
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
	if !id.ScopePermits(entity.PolicyObjectSSO, act) {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectSSO, act)
	if err != nil {
		return entity.Identity{}, entity.Workspace{}, err
	}
	if !allowed {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
	}
	return id, ws, nil
}

func (s *srv) GetConfig(ctx context.Context, workspaceSlug string) (entity.SSOConnection, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead)
	if err != nil {
		return entity.SSOConnection{}, err
	}
	conn, err := s.connections.Get(ctx, ws.ID)
	if err != nil {
		if errors.Is(err, entity.ErrSSONotConfigured) {
			return entity.SSOConnection{WorkspaceID: ws.ID, Mode: entity.SSOModeOIDC, Scopes: entity.SSODefaultScopes}, nil
		}
		return entity.SSOConnection{}, err
	}
	return conn, nil
}

func (s *srv) SaveConfig(ctx context.Context, workspaceSlug string, in entity.SSOConfigInput) (entity.SSOConnection, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite)
	if err != nil {
		return entity.SSOConnection{}, err
	}
	if len(in.Scopes) == 0 {
		in.Scopes = entity.SSODefaultScopes
	}
	in.AllowedEmailDomains = entity.NormalizeDomains(in.AllowedEmailDomains)
	if err := in.Validate(); err != nil {
		return entity.SSOConnection{}, err
	}
	var conn entity.SSOConnection
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		conn, err = s.connections.Save(ctx, ws.ID, in)
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, entity.AuditEvent{
			WorkspaceID: ws.ID, ActorType: entity.AuditActorUser, ActorUserID: actor.UserID,
			ActorLabel: actor.Label, Action: entity.ActionSSOUpdated, Target: string(in.Mode), IP: actor.IP,
		})
	})
	if err != nil {
		return entity.SSOConnection{}, err
	}
	return conn, nil
}

func (s *srv) StartLogin(ctx context.Context, workspaceSlug string) (string, error) {
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return "", err
	}
	conn, err := s.connections.Get(ctx, ws.ID)
	if err != nil {
		return "", err
	}
	if conn.Mode != entity.SSOModeOIDC {
		return "", entity.ErrSSOInvalid
	}
	if !conn.Enabled {
		return "", entity.ErrSSONotEnabled
	}
	secret, err := s.connections.ClientSecret(ctx, ws.ID)
	if err != nil {
		return "", err
	}
	provider, err := s.provider(ctx, conn.Issuer)
	if err != nil {
		return "", err
	}
	state, err := entity.GenerateToken(entity.SSOStateTokenLen)
	if err != nil {
		return "", err
	}
	nonce, err := entity.GenerateToken(entity.SSOStateTokenLen)
	if err != nil {
		return "", err
	}
	verifier := oauth2.GenerateVerifier()
	if err := s.states.Store(ctx, state, entity.SSOState{
		WorkspaceID: ws.ID, ConnectionID: conn.ID, Nonce: nonce, Verifier: verifier,
	}, entity.SSOStateTTL); err != nil {
		return "", err
	}
	cfg := s.oauthConfig(conn, secret, provider, workspaceSlug)
	return cfg.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)), nil
}

func (s *srv) CompleteLogin(ctx context.Context, workspaceSlug, code, state, ip, userAgent string) (entity.LoginResult, error) {
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return entity.LoginResult{}, err
	}
	st, err := s.states.Consume(ctx, state)
	if err != nil {
		return entity.LoginResult{}, err
	}
	if st.WorkspaceID != ws.ID {
		return entity.LoginResult{}, entity.ErrSSOStateInvalid
	}
	conn, err := s.connections.Get(ctx, ws.ID)
	if err != nil {
		return entity.LoginResult{}, err
	}
	if conn.ID != st.ConnectionID {
		return entity.LoginResult{}, entity.ErrSSOStateInvalid
	}
	secret, err := s.connections.ClientSecret(ctx, ws.ID)
	if err != nil {
		return entity.LoginResult{}, err
	}
	provider, err := s.provider(ctx, conn.Issuer)
	if err != nil {
		return entity.LoginResult{}, err
	}
	cfg := s.oauthConfig(conn, secret, provider, workspaceSlug)
	token, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(st.Verifier))
	if err != nil {
		return entity.LoginResult{}, fmt.Errorf("%w: %v", entity.ErrSSOExchange, err)
	}
	rawID, ok := token.Extra("id_token").(string)
	if !ok || rawID == "" {
		return entity.LoginResult{}, entity.ErrSSOExchange
	}
	idToken, err := provider.Verifier(&oidc.Config{ClientID: conn.ClientID}).Verify(ctx, rawID)
	if err != nil {
		return entity.LoginResult{}, fmt.Errorf("%w: %v", entity.ErrSSOExchange, err)
	}
	if idToken.Nonce != st.Nonce {
		return entity.LoginResult{}, entity.ErrSSOStateInvalid
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return entity.LoginResult{}, entity.ErrSSOExchange
	}
	email := entity.NormalizeEmail(claims.Email)
	if email == "" {
		return entity.LoginResult{}, entity.ErrSSOEmailMissing
	}
	name := strings.TrimSpace(claims.Name)
	if name == "" {
		name = email
	}

	var (
		user     entity.User
		assigned bool
	)
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		user, assigned, err = s.provision(ctx, ws, conn, idToken.Subject, email, name)
		return err
	})
	if err != nil {
		if assigned {
			if cErr := s.policy.RemoveRole(context.WithoutCancel(ctx), user.ID, ws.ID); cErr != nil {
				logger.From(ctx).ErrorContext(ctx, "sso provisioning compensation failed", "error", cErr, "user_id", user.ID, "workspace_id", ws.ID)
			}
		}
		return entity.LoginResult{}, err
	}
	return s.completeSession(ctx, ws, user, ip, userAgent)
}

func (s *srv) provision(ctx context.Context, ws entity.Workspace, conn entity.SSOConnection, subject, email, name string) (entity.User, bool, error) {
	ui, idErr := s.identities.GetBySubject(ctx, conn.ID, subject)
	if idErr == nil {
		user, err := s.users.GetByID(ctx, ui.UserID)
		if err != nil {
			return entity.User{}, false, err
		}
		assigned, err := s.ensureMembership(ctx, ws, user, email, conn)
		return user, assigned, err
	}
	if !errors.Is(idErr, entity.ErrUserIdentityNotFound) {
		return entity.User{}, false, idErr
	}

	existing, gErr := s.users.GetByEmail(ctx, email)
	if gErr == nil {
		assigned, err := s.ensureMembership(ctx, ws, existing, email, conn)
		if err != nil {
			return entity.User{}, assigned, err
		}
		if err := s.identities.Create(ctx, existing.ID, conn.ID, subject, email); err != nil {
			return entity.User{}, assigned, err
		}
		return existing, assigned, nil
	}
	if !errors.Is(gErr, entity.ErrUserNotFound) {
		return entity.User{}, false, gErr
	}

	if !conn.JITProvisioning {
		return entity.User{}, false, entity.ErrSSOProvisioningDisabled
	}
	if !entity.EmailDomainAllowed(email, conn.AllowedEmailDomains) {
		return entity.User{}, false, entity.ErrSSODomainNotAllowed
	}
	user, err := s.users.CreateSSO(ctx, email, name)
	if err != nil {
		return entity.User{}, false, err
	}
	if err := s.members.Create(ctx, ws.ID, user.ID, entity.MemberStatusActive); err != nil {
		return entity.User{}, false, err
	}
	if err := s.policy.AssignRole(ctx, user.ID, ws.ID, entity.RoleMember); err != nil {
		return user, true, err
	}
	if err := s.audit.Create(ctx, s.joinEvent(ws.ID, user)); err != nil {
		return user, true, err
	}
	if err := s.identities.Create(ctx, user.ID, conn.ID, subject, email); err != nil {
		return user, true, err
	}
	return user, true, nil
}

func (s *srv) ensureMembership(ctx context.Context, ws entity.Workspace, user entity.User, email string, conn entity.SSOConnection) (bool, error) {
	m, err := s.members.Get(ctx, ws.ID, user.ID)
	if err == nil {
		switch m.Status {
		case entity.MemberStatusActive:
			return false, nil
		case entity.MemberStatusInvited:
			return false, s.members.UpdateStatus(ctx, ws.ID, user.ID, entity.MemberStatusActive)
		case entity.MemberStatusDeactivated:
			return false, entity.ErrMemberDeactivated
		default:
			return false, nil
		}
	}
	if !errors.Is(err, entity.ErrMemberNotFound) {
		return false, err
	}
	if !conn.JITProvisioning {
		return false, entity.ErrSSOProvisioningDisabled
	}
	if !entity.EmailDomainAllowed(email, conn.AllowedEmailDomains) {
		return false, entity.ErrSSODomainNotAllowed
	}
	if err := s.members.Create(ctx, ws.ID, user.ID, entity.MemberStatusActive); err != nil {
		return false, err
	}
	if err := s.policy.AssignRole(ctx, user.ID, ws.ID, entity.RoleMember); err != nil {
		return true, err
	}
	if err := s.audit.Create(ctx, s.joinEvent(ws.ID, user)); err != nil {
		return true, err
	}
	return true, nil
}

func (s *srv) completeSession(ctx context.Context, ws entity.Workspace, user entity.User, ip, userAgent string) (entity.LoginResult, error) {
	token, err := entity.GenerateToken(entity.SessionTokenLength)
	if err != nil {
		return entity.LoginResult{}, err
	}
	sess, err := s.sessions.Create(ctx, user.ID, entity.HashToken(token), ip, userAgent, time.Now().Add(s.cfg.SessionAbsoluteTTL))
	if err != nil {
		return entity.LoginResult{}, err
	}
	if err := s.audit.Create(ctx, entity.AuditEvent{
		WorkspaceID: ws.ID, ActorType: entity.AuditActorUser, ActorUserID: user.ID,
		ActorLabel: user.Name, Action: entity.ActionAuthLogin, Target: user.Email, IP: ip,
	}); err != nil {
		return entity.LoginResult{}, err
	}
	return entity.LoginResult{Outcome: entity.LoginOutcomeOK, Session: sess, Token: token, User: user}, nil
}

func (s *srv) joinEvent(workspaceID string, user entity.User) entity.AuditEvent {
	return entity.AuditEvent{
		WorkspaceID: workspaceID, ActorType: entity.AuditActorUser, ActorUserID: user.ID,
		ActorLabel: user.Name, Action: entity.ActionMemberJoined, Target: user.Email,
	}
}

func (s *srv) oauthConfig(conn entity.SSOConnection, secret string, provider *oidc.Provider, workspaceSlug string) oauth2.Config {
	return oauth2.Config{
		ClientID:     conn.ClientID,
		ClientSecret: secret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  s.redirectURL(workspaceSlug),
		Scopes:       conn.Scopes,
	}
}

func (s *srv) redirectURL(workspaceSlug string) string {
	return strings.TrimRight(s.cfg.BaseURL, "/") + "/v1/auth/sso/" + workspaceSlug + "/callback"
}

func (s *srv) provider(ctx context.Context, issuer string) (*oidc.Provider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.providers[issuer]; ok {
		return p, nil
	}
	p, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	s.providers[issuer] = p
	return p, nil
}
