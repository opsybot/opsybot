package entity

import (
	"errors"
	"time"
)

type RateScope string

const (
	RateScopeLogin         RateScope = "login"
	RateScopeSignup        RateScope = "signup"
	RateScopeSlugCheck     RateScope = "slug_check"
	RateScopePasswordReset RateScope = "password_reset"
	RateScopeSSO           RateScope = "sso"
	RateScopeIngest        RateScope = "ingest"
)

const RateLimitFailClosedRetry = 5 * time.Second

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
