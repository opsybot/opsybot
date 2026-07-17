package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/service"
)

type rateRule struct {
	method string
	path   string
	scope  entity.RateScope
}

var rateRules = []rateRule{
	{http.MethodPost, "/v1/auth/login", entity.RateScopeLogin},
	{http.MethodPost, "/v1/auth/setup", entity.RateScopeLogin},
	{http.MethodPost, "/v1/auth/two-factor/verify", entity.RateScopeLogin},
	{http.MethodPost, "/v1/auth/two-factor/recovery", entity.RateScopeLogin},
	{http.MethodPost, "/v1/auth/invite/accept", entity.RateScopeLogin},
	{http.MethodPost, "/v1/auth/password/forgot", entity.RateScopePasswordReset},
	{http.MethodPost, "/v1/auth/password/reset", entity.RateScopePasswordReset},
}

func rateScope(r *http.Request) (entity.RateScope, bool) {
	for _, rule := range rateRules {
		if rule.method == r.Method && rule.path == r.URL.Path {
			return rule.scope, true
		}
	}
	if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/auth/sso/") && strings.HasSuffix(r.URL.Path, "/start") {
		return entity.RateScopeSSO, true
	}
	return "", false
}

func RateLimit(limiter service.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scope, ok := rateScope(r)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			key := entity.RequestInfoFrom(r.Context()).IP
			if key == "" {
				key = r.RemoteAddr
			}
			res, err := limiter.Allow(r.Context(), scope, key)
			if err == nil && !res.Allowed {
				writeRateLimited(w, res.RetryAfter)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeRateLimited(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := max(int(retryAfter.Seconds()), 1)
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "https://opsybot.dev/problems/rate-limited",
		"title":  "Too many requests",
		"status": http.StatusTooManyRequests,
		"detail": "You've made too many attempts. Wait a moment and try again.",
	})
}
