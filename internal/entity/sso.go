package entity

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

type SSOMode string

const (
	SSOModeOIDC SSOMode = "oidc"
	SSOModeSAML SSOMode = "saml"
)

const (
	SSOStateTTL       = 10 * time.Minute
	SSOStateTokenLen  = 32
	SSOSPCertValidity = 10 * 365 * 24 * time.Hour
)

var SSODefaultScopes = []string{"openid", "email", "profile"}

type SSOConnection struct {
	ID                  string
	WorkspaceID         string
	Mode                SSOMode
	Issuer              string
	ClientID            string
	HasClientSecret     bool
	Scopes              []string
	SAMLMetadataURL     string
	Enabled             bool
	Required            bool
	JITProvisioning     bool
	AllowedEmailDomains []string
	UpdatedAt           time.Time
}

type SSOConfigInput struct {
	Mode                SSOMode
	Issuer              string
	ClientID            string
	ClientSecret        string
	ClearClientSecret   bool
	Scopes              []string
	SAMLMetadataURL     string
	Enabled             bool
	Required            bool
	JITProvisioning     bool
	AllowedEmailDomains []string
}

type SSOClaims struct {
	Subject       string
	Email         string
	Name          string
	EmailVerified bool
}

type UserIdentity struct {
	ID           string
	UserID       string
	ConnectionID string
	Subject      string
	Email        string
}

type SSOState struct {
	WorkspaceID  string
	ConnectionID string
	Nonce        string
	Verifier     string
}

var (
	ErrSSONotConfigured        = errors.New("sso not configured")
	ErrSSONotEnabled           = errors.New("sso not enabled")
	ErrSSOInvalid              = errors.New("sso invalid")
	ErrSSOUnavailable          = errors.New("sso unavailable")
	ErrSSOStateInvalid         = errors.New("sso state invalid")
	ErrSSOExchange             = errors.New("sso exchange failed")
	ErrSSOEmailMissing         = errors.New("sso email missing")
	ErrSSOEmailUnverified      = errors.New("sso email unverified")
	ErrSSOProvisioningDisabled = errors.New("sso provisioning disabled")
	ErrSSODomainNotAllowed     = errors.New("sso domain not allowed")
	ErrUserIdentityNotFound    = errors.New("user identity not found")
	ErrUserIdentityExists      = errors.New("user identity exists")
)

func (m SSOMode) Validate() error {
	switch m {
	case SSOModeOIDC, SSOModeSAML:
		return nil
	default:
		return ErrSSOInvalid
	}
}

func (in SSOConfigInput) Validate() error {
	if err := in.Mode.Validate(); err != nil {
		return err
	}
	if in.Mode == SSOModeOIDC {
		if !validHTTPSURL(in.Issuer) || strings.TrimSpace(in.ClientID) == "" {
			return ErrSSOInvalid
		}
	}
	if in.Mode == SSOModeSAML && !validHTTPSURL(in.SAMLMetadataURL) {
		return ErrSSOInvalid
	}
	for _, d := range in.AllowedEmailDomains {
		if strings.TrimSpace(d) == "" || strings.ContainsAny(d, "@ ") {
			return ErrSSOInvalid
		}
	}
	return nil
}

func validHTTPSURL(s string) bool {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil || u.Host == "" {
		return false
	}
	if u.Scheme == "https" {
		return true
	}
	return u.Scheme == "http" && isLoopbackHost(u.Hostname())
}

func isLoopbackHost(host string) bool {
	return host == "localhost" || host == "::1" || strings.HasPrefix(host, "127.")
}

func EmailDomain(email string) string {
	at := strings.LastIndexByte(email, '@')
	if at < 0 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(email[at+1:]))
}

func EmailDomainAllowed(email string, domains []string) bool {
	if len(domains) == 0 {
		return true
	}
	d := EmailDomain(email)
	for _, allowed := range domains {
		if d != "" && strings.EqualFold(strings.TrimSpace(allowed), d) {
			return true
		}
	}
	return false
}

func NormalizeDomains(domains []string) []string {
	seen := make(map[string]struct{}, len(domains))
	out := make([]string, 0, len(domains))
	for _, d := range domains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == "" {
			continue
		}
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}
	return out
}
