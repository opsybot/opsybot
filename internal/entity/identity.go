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

func (i Identity) Subject() string {
	if i.Kind == IdentityKindAPIKey && i.KeyKind == KeyKindWorkspace {
		return "wsagent:" + i.WorkspaceID
	}
	return "user:" + i.UserID
}

func (i Identity) ScopeAllows(scope Scope) bool {
	if i.Kind != IdentityKindAPIKey {
		return true
	}
	return slices.Contains(i.Scopes, scope)
}
