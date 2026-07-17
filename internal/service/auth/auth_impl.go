package auth

import (
	"context"
	"errors"
	"fmt"
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
	users      repository.User
	workspaces repository.Workspace
	members    repository.Member
	sessions   repository.Session
	policy     repository.Policy
}

func New(
	cfg config.Auth,
	tx repository.Transactor,
	lock repository.Lock,
	users repository.User,
	workspaces repository.Workspace,
	members repository.Member,
	sessions repository.Session,
	policy repository.Policy,
) service.Auth {
	return &srv{cfg: cfg, tx: tx, lock: lock, users: users, workspaces: workspaces, members: members, sessions: sessions, policy: policy}
}

func (s *srv) SetupRequired(ctx context.Context) (bool, error) {
	exists, err := s.users.ExistsAny(ctx)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (s *srv) Setup(ctx context.Context, in entity.Setup, ip, userAgent string) (entity.SetupResult, error) {
	if err := in.Validate(); err != nil {
		return entity.SetupResult{}, err
	}
	hash, err := entity.HashPassword(in.Password)
	if err != nil {
		return entity.SetupResult{}, err
	}

	var (
		user entity.User
		ws   entity.Workspace
	)
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.lock.Instance(ctx); err != nil {
			return err
		}
		exists, err := s.users.ExistsAny(ctx)
		if err != nil {
			return err
		}
		if exists {
			return entity.ErrSetupAlreadyDone
		}
		user, err = s.users.Create(ctx, entity.NewUser{Email: in.Email, Name: in.UserName, Timezone: in.Timezone}, hash)
		if err != nil {
			return err
		}
		ws, err = s.workspaces.Create(ctx, entity.NewWorkspace{
			Slug: in.WorkspaceSlug, Name: in.WorkspaceName, Timezone: in.Timezone,
		}, user.ID)
		if err != nil {
			return err
		}
		if err := s.members.Create(ctx, ws.ID, user.ID, entity.MemberStatusActive); err != nil {
			return err
		}
		if err := s.policy.SeedWorkspace(ctx, ws.ID); err != nil {
			return err
		}
		return s.policy.AssignRole(ctx, user.ID, ws.ID, entity.RoleAdmin)
	})
	if err != nil {
		if compErr := s.policy.RemoveRole(context.WithoutCancel(ctx), user.ID, ws.ID); compErr != nil && user.ID != "" {
			logger.From(ctx).ErrorContext(ctx, "setup casbin compensation failed", "error", compErr, "user_id", user.ID, "workspace_id", ws.ID)
		}
		return entity.SetupResult{}, err
	}

	sess, token, err := s.issueSession(ctx, user.ID, ip, userAgent, true)
	if err != nil {
		return entity.SetupResult{}, err
	}
	return entity.SetupResult{Workspace: ws, Session: sess, Token: token, User: user}, nil
}

func (s *srv) Login(ctx context.Context, in entity.LoginInput) (entity.LoginResult, error) {
	user, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			_, _ = entity.HashPassword(in.Password)
			return entity.LoginResult{}, entity.ErrInvalidCredentials
		}
		return entity.LoginResult{}, err
	}

	hash, err := s.users.PasswordHash(ctx, user.ID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNoPassword) {
			return entity.LoginResult{}, entity.ErrSSORequired
		}
		return entity.LoginResult{}, err
	}
	if err := entity.VerifyPassword(hash, in.Password); err != nil {
		return entity.LoginResult{}, entity.ErrInvalidCredentials
	}

	workspaces, err := s.workspaces.ListActiveByUser(ctx, user.ID)
	if err != nil {
		return entity.LoginResult{}, err
	}
	if len(workspaces) == 0 {
		return entity.LoginResult{}, entity.ErrUserDeactivated
	}

	sess, token, err := s.issueSession(ctx, user.ID, in.IP, in.UserAgent, in.Remember)
	if err != nil {
		return entity.LoginResult{}, err
	}
	return entity.LoginResult{Outcome: entity.LoginOutcomeOK, Session: sess, Token: token, User: user}, nil
}

func (s *srv) Logout(ctx context.Context) error {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.ErrUnauthenticated
	}
	if id.Kind != entity.IdentityKindSession {
		return entity.ErrForbidden
	}
	return s.sessions.Delete(ctx, id.SessionID)
}

func (s *srv) Resolve(ctx context.Context, token string) (entity.Identity, error) {
	sess, err := s.sessions.GetByTokenHash(ctx, entity.HashToken(token))
	if err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			return entity.Identity{}, entity.ErrUnauthenticated
		}
		return entity.Identity{}, err
	}
	now := time.Now()
	if now.After(sess.ExpiresAt) {
		return entity.Identity{}, entity.ErrSessionExpired
	}
	if now.Sub(sess.LastSeenAt) > s.cfg.SessionTouchWindow {
		idle := s.cfg.SessionIdleTTL
		if err := s.sessions.Touch(ctx, sess.ID, now, now.Add(idle)); err != nil {
			return entity.Identity{}, err
		}
		if err := s.users.TouchLastActive(ctx, sess.UserID); err != nil {
			return entity.Identity{}, err
		}
	}
	user, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return entity.Identity{}, err
	}
	return entity.Identity{
		Kind:      entity.IdentityKindSession,
		UserID:    sess.UserID,
		SessionID: sess.ID,
		Label:     user.Name,
	}, nil
}

func (s *srv) Profile(ctx context.Context) (entity.User, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.User{}, entity.ErrUnauthenticated
	}
	return s.users.GetByID(ctx, id.UserID)
}

func (s *srv) issueSession(ctx context.Context, userID, ip, userAgent string, remember bool) (entity.Session, string, error) {
	token, err := entity.GenerateToken(entity.SessionTokenLength)
	if err != nil {
		return entity.Session{}, "", err
	}
	ttl := s.cfg.SessionBrowserTTL
	if remember {
		ttl = s.cfg.SessionAbsoluteTTL
	}
	sess, err := s.sessions.Create(ctx, userID, entity.HashToken(token), ip, userAgent, time.Now().Add(ttl))
	if err != nil {
		return entity.Session{}, "", fmt.Errorf("issue session: %w", err)
	}
	return sess, token, nil
}
