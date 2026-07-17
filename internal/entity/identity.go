package entity

import (
	"context"
	"errors"
	"slices"
)

type IdentityKind string

const (
	IdentityKindSession IdentityKind = "session"
	IdentityKindAPIKey  IdentityKind = "api_key"
)

type Identity struct {
	Kind        IdentityKind
	UserID      string
	SessionID   string
	APIKeyID    string
	KeyKind     KeyKind
	WorkspaceID string
	Label       string
	IP          string
	UserAgent   string
	Scopes      []Scope
}

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
)

type identityCtxKey struct{}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityCtxKey{}, id)
}

func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityCtxKey{}).(Identity)
	return id, ok
}

type RequestInfo struct {
	IP        string
	UserAgent string
}

type requestInfoCtxKey struct{}

func WithRequestInfo(ctx context.Context, info RequestInfo) context.Context {
	return context.WithValue(ctx, requestInfoCtxKey{}, info)
}

func RequestInfoFrom(ctx context.Context) RequestInfo {
	info, _ := ctx.Value(requestInfoCtxKey{}).(RequestInfo)
	return info
}

const PendingCookieName = "opsybot_2fa"

type pendingTokenCtxKey struct{}

func WithPendingToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, pendingTokenCtxKey{}, token)
}

func PendingTokenFrom(ctx context.Context) string {
	token, _ := ctx.Value(pendingTokenCtxKey{}).(string)
	return token
}

func (i Identity) Subject() string {
	if i.Kind == IdentityKindAPIKey && i.KeyKind == KeyKindWorkspace {
		return "wsagent:" + i.WorkspaceID
	}
	return "user:" + i.UserID
}

func (i Identity) ScopePermits(obj PolicyObject, act PolicyAction) bool {
	if i.Kind != IdentityKindAPIKey {
		return true
	}
	scope, ok := ScopeFor(obj, act)
	if !ok {
		return false
	}
	return slices.Contains(i.Scopes, scope)
}
