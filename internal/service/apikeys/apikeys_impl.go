package apikeys

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx         repository.Transactor
	workspaces repository.Workspace
	members    repository.Member
	keys       repository.APIKey
	policy     repository.Policy
	audit      repository.Audit
}

func New(
	tx repository.Transactor,
	workspaces repository.Workspace,
	members repository.Member,
	keys repository.APIKey,
	policy repository.Policy,
	audit repository.Audit,
) service.APIKeys {
	return &srv{tx: tx, workspaces: workspaces, members: members, keys: keys, policy: policy, audit: audit}
}

func (s *srv) member(ctx context.Context, workspaceSlug string) (entity.Identity, entity.Workspace, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok {
		return entity.Identity{}, entity.Workspace{}, entity.ErrUnauthenticated
	}
	if id.Kind != entity.IdentityKindSession {
		return entity.Identity{}, entity.Workspace{}, entity.ErrForbidden
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
	return id, ws, nil
}

func (s *srv) List(ctx context.Context, workspaceSlug string) (entity.APIKeyList, error) {
	id, ws, err := s.member(ctx, workspaceSlug)
	if err != nil {
		return entity.APIKeyList{}, err
	}
	personal, err := s.keys.ListByOwner(ctx, ws.ID, id.UserID)
	if err != nil {
		return entity.APIKeyList{}, err
	}
	out := entity.APIKeyList{Personal: personal}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, entity.PolicyObjectKeys, entity.PolicyActionRead)
	if err != nil {
		return entity.APIKeyList{}, err
	}
	if allowed {
		out.Workspace, err = s.keys.ListWorkspaceKeys(ctx, ws.ID)
		if err != nil {
			return entity.APIKeyList{}, err
		}
	}
	return out, nil
}

func (s *srv) Create(ctx context.Context, workspaceSlug string, in entity.NewAPIKey) (entity.APIKey, string, error) {
	id, ws, err := s.member(ctx, workspaceSlug)
	if err != nil {
		return entity.APIKey{}, "", err
	}
	if err := in.Validate(); err != nil {
		return entity.APIKey{}, "", err
	}
	obj := entity.PolicyObjectPersonalKeys
	if in.Kind == entity.KeyKindWorkspace {
		obj = entity.PolicyObjectKeys
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), ws.ID, obj, entity.PolicyActionWrite)
	if err != nil {
		return entity.APIKey{}, "", err
	}
	if !allowed {
		return entity.APIKey{}, "", entity.ErrForbidden
	}

	secret, hint, hash, err := entity.NewAPIKeySecret()
	if err != nil {
		return entity.APIKey{}, "", err
	}
	rec := entity.APIKeyRecord{
		WorkspaceID: ws.ID,
		Kind:        in.Kind,
		CreatedBy:   id.UserID,
		Name:        strings.TrimSpace(in.Name),
		TokenHash:   hash,
		TokenHint:   hint,
		Scopes:      dedupeScopes(in.Scopes),
	}
	if in.Kind == entity.KeyKindPersonal {
		rec.OwnerUserID = id.UserID
	}

	var key entity.APIKey
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		key, err = s.keys.Create(ctx, rec)
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(id, ws.ID, entity.ActionKeyCreated, key.Name))
	})
	if err != nil {
		return entity.APIKey{}, "", err
	}
	return key, secret, nil
}

func (s *srv) Revoke(ctx context.Context, workspaceSlug, keyID string) error {
	id, ws, err := s.member(ctx, workspaceSlug)
	if err != nil {
		return err
	}
	key, err := s.keys.GetByID(ctx, ws.ID, keyID)
	if err != nil {
		return err
	}
	if !key.RevokedAt.IsZero() {
		return entity.ErrAPIKeyRevoked
	}
	if err := s.authorizeMutation(ctx, id, ws.ID, key); err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.keys.Revoke(ctx, ws.ID, keyID); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(id, ws.ID, entity.ActionKeyRevoked, key.Name))
	})
}

func (s *srv) authorizeMutation(ctx context.Context, id entity.Identity, workspaceID string, key entity.APIKey) error {
	if key.Kind == entity.KeyKindPersonal {
		if key.OwnerUserID != id.UserID {
			return entity.ErrForbidden
		}
		return nil
	}
	allowed, err := s.policy.Allowed(ctx, id.Subject(), workspaceID, entity.PolicyObjectKeys, entity.PolicyActionWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return entity.ErrForbidden
	}
	return nil
}

func (s *srv) Resolve(ctx context.Context, secret string) (entity.Identity, error) {
	key, err := s.keys.GetByTokenHash(ctx, entity.HashToken(secret))
	if err != nil {
		if errors.Is(err, entity.ErrAPIKeyNotFound) {
			return entity.Identity{}, entity.ErrUnauthenticated
		}
		return entity.Identity{}, err
	}
	if !key.RevokedAt.IsZero() {
		return entity.Identity{}, entity.ErrUnauthenticated
	}
	now := time.Now()
	if now.Sub(key.LastUsedAt) > entity.APIKeyTouchWindow {
		if err := s.keys.TouchLastUsed(ctx, key.ID, now); err != nil {
			return entity.Identity{}, err
		}
	}
	id := entity.Identity{
		Kind:        entity.IdentityKindAPIKey,
		APIKeyID:    key.ID,
		KeyKind:     key.Kind,
		WorkspaceID: key.WorkspaceID,
		Scopes:      key.Scopes,
		Label:       key.Name,
	}
	if key.Kind == entity.KeyKindPersonal {
		id.UserID = key.OwnerUserID
	}
	return id, nil
}

func dedupeScopes(in []entity.Scope) []entity.Scope {
	seen := make(map[entity.Scope]struct{}, len(in))
	out := make([]entity.Scope, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
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
