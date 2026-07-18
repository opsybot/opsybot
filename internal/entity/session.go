package entity

import (
	"errors"
	"time"
)

const (
	SessionTokenLength     = 32
	SessionIdleTTL         = 72 * time.Hour
	SessionAbsoluteTTL     = 720 * time.Hour
	SessionBrowserTTL      = 24 * time.Hour
	SessionTouchInterval   = 5 * time.Minute
	PendingTwoFactorTTL    = 5 * time.Minute
	PendingTwoFactorLength = 32
)

type Session struct {
	ID                string
	UserID            string
	IP                string
	UserAgent         string
	ExpiresAt         time.Time
	AbsoluteExpiresAt time.Time
	LastSeenAt        time.Time
	CreatedAt         time.Time
}

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type LoginOutcome string

const (
	LoginOutcomeOK          LoginOutcome = "ok"
	LoginOutcomeTwoFactor   LoginOutcome = "two_factor_required"
	LoginInvalidCredentials              = "invalid"
	LoginDeactivated                     = "deactivated"
	LoginSSORequired                     = "sso_required"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDeactivated    = errors.New("user deactivated")
	ErrSSORequired        = errors.New("sso required")
)

type LoginInput struct {
	Email     string
	Password  string
	IP        string
	UserAgent string
	Remember  bool
}

type LoginResult struct {
	Outcome      LoginOutcome
	Session      Session
	Token        string
	User         User
	PendingToken string
}

type SetupResult struct {
	Workspace Workspace
	Session   Session
	Token     string
	User      User
}

type PendingTwoFactor struct {
	UserID    string
	Remember  bool
	IP        string
	UserAgent string
}

const PendingTwoFactorMaxAttempts = 5
