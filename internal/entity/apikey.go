package entity

import (
	"errors"
	"slices"
	"strings"
	"time"
)

type KeyKind string

const (
	KeyKindPersonal  KeyKind = "personal"
	KeyKindWorkspace KeyKind = "workspace"
)

type Scope string

const (
	ScopeIncidentsRead  Scope = "incidents:read"
	ScopeIncidentsWrite Scope = "incidents:write"
	ScopeAlertsRead     Scope = "alerts:read"
	ScopeAlertsWrite    Scope = "alerts:write"
	ScopeSchedulesWrite Scope = "schedules:write"
	ScopePoliciesWrite  Scope = "policies:write"
	ScopeConfigWrite    Scope = "config:write"
	ScopeAuditRead      Scope = "audit:read"
)

const (
	APIKeySecretPrefix     = "osk"
	APIKeySecretHexLength  = 24
	APIKeyNameMaxLength    = 60
	APIKeyHintSuffixLength = 4
)

var AllScopes = []Scope{
	ScopeIncidentsRead, ScopeIncidentsWrite, ScopeAlertsRead, ScopeAlertsWrite,
	ScopeSchedulesWrite, ScopePoliciesWrite, ScopeConfigWrite, ScopeAuditRead,
}

type APIKey struct {
	ID          string
	WorkspaceID string
	OwnerUserID string
	CreatedBy   string
	Name        string
	TokenHint   string
	Kind        KeyKind
	Scopes      []Scope
	LastUsedAt  time.Time
	RevokedAt   time.Time
	CreatedAt   time.Time
}

type NewAPIKey struct {
	Name   string
	Kind   KeyKind
	Scopes []Scope
}

var (
	ErrAPIKeyNotFound     = errors.New("api key not found")
	ErrAPIKeyRevoked      = errors.New("api key revoked")
	ErrAPIKeyInvalidScope = errors.New("api key invalid scope")
	ErrAPIKeyInvalidKind  = errors.New("api key invalid kind")
	ErrAPIKeyInvalidName  = errors.New("api key invalid name")
)

func (k KeyKind) Validate() error {
	switch k {
	case KeyKindPersonal, KeyKindWorkspace:
		return nil
	default:
		return ErrAPIKeyInvalidKind
	}
}

func ScopeValid(s Scope) bool {
	return slices.Contains(AllScopes, s)
}

func (n NewAPIKey) Validate() error {
	name := strings.TrimSpace(n.Name)
	if name == "" || len(name) > APIKeyNameMaxLength {
		return ErrAPIKeyInvalidName
	}
	if err := n.Kind.Validate(); err != nil {
		return err
	}
	if len(n.Scopes) == 0 {
		return ErrAPIKeyInvalidScope
	}
	for _, s := range n.Scopes {
		if !ScopeValid(s) {
			return ErrAPIKeyInvalidScope
		}
	}
	return nil
}

func KindAbbrev(k KeyKind) string {
	if k == KeyKindWorkspace {
		return "wo"
	}
	return "pe"
}
