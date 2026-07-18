package middleware

import (
	"net"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
)

func ClientInfo(trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := entity.RequestInfo{IP: clientIP(r, trustProxy), UserAgent: r.UserAgent()}
			ctx := entity.WithRequestInfo(r.Context(), info)
			if cookie, err := r.Cookie(entity.PendingCookieName); err == nil {
				ctx = entity.WithPendingToken(ctx, cookie.Value)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := splitAndTrim(xff)
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func splitAndTrim(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			for len(part) > 0 && part[0] == ' ' {
				part = part[1:]
			}
			for len(part) > 0 && part[len(part)-1] == ' ' {
				part = part[:len(part)-1]
			}
			if part != "" {
				out = append(out, part)
			}
			start = i + 1
		}
	}
	return out
}
