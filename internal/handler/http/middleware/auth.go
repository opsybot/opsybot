package middleware

import (
	"encoding/json"
	"errors"
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
	{http.MethodPost, "/v1/auth/signup"},
	{http.MethodGet, "/v1/auth/slug-available"},
	{http.MethodPost, "/v1/auth/login"},
	{http.MethodPost, "/v1/auth/invite/preview"},
	{http.MethodPost, "/v1/auth/invite/accept"},
	{http.MethodPost, "/v1/auth/two-factor/verify"},
	{http.MethodPost, "/v1/auth/two-factor/recovery"},
	{http.MethodPost, "/v1/auth/password/forgot"},
	{http.MethodPost, "/v1/auth/password/reset"},
}

func isPublic(r *http.Request) bool {
	for _, pr := range publicRoutes {
		if pr.method == r.Method && pr.path == r.URL.Path {
			return true
		}
	}
	return strings.HasPrefix(r.URL.Path, "/v1/auth/sso/") ||
		strings.HasPrefix(r.URL.Path, "/v1/ingest/") ||
		strings.HasPrefix(r.URL.Path, "/v1/channels/verify/") ||
		strings.HasPrefix(r.URL.Path, "/v1/chat/") ||
		strings.HasPrefix(r.URL.Path, "/v1/act/") ||
		(r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/oncall/feed/"))
}

func Authn(auth service.Auth, keys service.APIKeys, cfg config.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublic(r) {
				next.ServeHTTP(w, r)
				return
			}
			id, err := resolveIdentity(r, auth, keys, cfg)
			if err != nil {
				writeUnauthorized(w, unauthorizedDetail(err))
				return
			}
			info := entity.RequestInfoFrom(r.Context())
			id.IP = info.IP
			id.UserAgent = info.UserAgent
			next.ServeHTTP(w, r.WithContext(entity.WithIdentity(r.Context(), id)))
		})
	}
}

func resolveIdentity(r *http.Request, auth service.Auth, keys service.APIKeys, cfg config.Auth) (entity.Identity, error) {
	if token, ok := bearerToken(r); ok {
		return keys.Resolve(r.Context(), token)
	}
	cookie, err := r.Cookie(cfg.CookieName)
	if err != nil {
		return entity.Identity{}, entity.ErrUnauthenticated
	}
	return auth.Resolve(r.Context(), cookie.Value)
}

func bearerToken(r *http.Request) (string, bool) {
	if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		if token = strings.TrimSpace(token); token != "" {
			return token, true
		}
	}
	return "", false
}

func unauthorizedDetail(err error) string {
	if errors.Is(err, entity.ErrUnauthenticated) {
		return "You're not signed in. This request had no valid credential. Sign in again to continue."
	}
	return "Your session is no longer valid. It may have expired or been revoked. Sign in again to continue."
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
