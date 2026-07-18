package entity

import (
	"errors"
	"slices"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	APIKeyTouchWindow      = 5 * time.Minute
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

type APIKeyRecord struct {
	WorkspaceID string
	Kind        KeyKind
	OwnerUserID string
	CreatedBy   string
	Name        string
	TokenHash   string
	TokenHint   string
	Scopes      []Scope
}

type APIKeyList struct {
	Personal  []APIKey
	Workspace []APIKey
}

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyRevoked  = errors.New("api key revoked")
)

func (k KeyKind) Validate() error {
	return keyKindField(k)
}

func ScopeValid(s Scope) bool {
	return slices.Contains(AllScopes, s)
}

func (n NewAPIKey) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Name, validation.By(keyNameField)),
		validation.Field(&n.Kind, validation.By(keyKindField)),
		validation.Field(&n.Scopes, validation.By(keyScopesField)),
	)
}

func KindAbbrev(k KeyKind) string {
	if k == KeyKindWorkspace {
		return "wo"
	}
	return "pe"
}

func NewAPIKeySecret() (secret, hint, hash string, err error) {
	raw, err := GenerateHexToken(APIKeySecretHexLength)
	if err != nil {
		return "", "", "", err
	}
	secret = APIKeySecretPrefix + "_" + raw
	hint = raw[len(raw)-APIKeyHintSuffixLength:]
	hash = HashToken(secret)
	return secret, hint, hash, nil
}
