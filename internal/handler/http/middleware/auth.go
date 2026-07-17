package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/service"
)

type publicRoute struct {
	method string
	path   string
}

var publicRoutes = []publicRoute{
	{http.MethodGet, "/v1/health"},
	{http.MethodGet, "/v1/auth/setup"},
	{http.MethodPost, "/v1/auth/setup"},
	{http.MethodPost, "/v1/auth/login"},
	{http.MethodPost, "/v1/auth/invite/preview"},
	{http.MethodPost, "/v1/auth/invite/accept"},
}

func isPublic(r *http.Request) bool {
	for _, pr := range publicRoutes {
		if pr.method == r.Method && pr.path == r.URL.Path {
			return true
		}
	}
	return strings.HasPrefix(r.URL.Path, "/v1/auth/sso/")
}

func Authn(auth service.Auth, cfg config.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublic(r) {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil {
				writeUnauthorized(w, "You're not signed in. This request had no valid session. Sign in again to continue.")
				return
			}
			id, err := auth.Resolve(r.Context(), cookie.Value)
			if err != nil {
				writeUnauthorized(w, "Your session is no longer valid. It may have expired or been revoked. Sign in again to continue.")
				return
			}
			info := entity.RequestInfoFrom(r.Context())
			id.IP = info.IP
			id.UserAgent = info.UserAgent
			next.ServeHTTP(w, r.WithContext(entity.WithIdentity(r.Context(), id)))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "about:blank",
		"title":  "Not authenticated",
		"status": http.StatusUnauthorized,
		"detail": detail,
	})
}
