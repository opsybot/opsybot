package ratelimiter

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	cfg     config.Auth
	limiter repository.RateLimiter
}

func New(cfg config.Auth, limiter repository.RateLimiter) service.RateLimiter {
	return &srv{cfg: cfg, limiter: limiter}
}

func (s *srv) Allow(ctx context.Context, scope entity.RateScope, key string) (entity.RateResult, error) {
	limit := s.limitFor(scope)
	if limit.Rate <= 0 {
		return entity.RateResult{Allowed: true}, nil
	}
	return s.limiter.Allow(ctx, string(scope)+":"+key, limit)
}

func (s *srv) limitFor(scope entity.RateScope) entity.RateLimit {
	switch scope {
	case entity.RateScopeLogin:
		return entity.RateLimit{Rate: s.cfg.RateLoginPerMin, Period: time.Minute, Burst: s.cfg.RateLoginPerMin}
	case entity.RateScopeSignup:
		return entity.RateLimit{Rate: s.cfg.RateSignupPerHour, Period: time.Hour, Burst: s.cfg.RateSignupPerHour}
	case entity.RateScopeSlugCheck:
		return entity.RateLimit{Rate: s.cfg.RateSlugCheckPerMin, Period: time.Minute, Burst: s.cfg.RateSlugCheckPerMin}
	case entity.RateScopePasswordReset:
		return entity.RateLimit{Rate: s.cfg.RateResetPerHour, Period: time.Hour, Burst: s.cfg.RateResetPerHour}
	case entity.RateScopeSSO:
		return entity.RateLimit{Rate: s.cfg.RateSSOPerMin, Period: time.Minute, Burst: s.cfg.RateSSOPerMin}
	case entity.RateScopeNotify:
		return entity.RateLimit{Rate: s.cfg.RateNotifyPerMin, Period: time.Minute, Burst: s.cfg.RateNotifyPerMin}
	case entity.RateScopeChannelTest:
		return entity.RateLimit{Rate: s.cfg.RateChannelTestPerHour, Period: time.Hour, Burst: s.cfg.RateChannelTestPerHour}
	default:
		return entity.RateLimit{}
	}
}
