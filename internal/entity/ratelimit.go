package entity

import (
	"errors"
	"time"
)

type RateScope string

const (
	RateScopeLogin         RateScope = "login"
	RateScopeSignup        RateScope = "signup"
	RateScopePasswordReset RateScope = "password_reset"
	RateScopeSSO           RateScope = "sso"
)

type RateLimit struct {
	Rate   int
	Period time.Duration
	Burst  int
}

type RateResult struct {
	Allowed    bool
	RetryAfter time.Duration
}

var ErrRateLimited = errors.New("rate limited")
