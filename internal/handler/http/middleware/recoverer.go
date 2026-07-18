package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/opsybot/opsybot/internal/pkg/logger"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil || rec == http.ErrAbortHandler {
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				return
			}
			logger.From(r.Context()).ErrorContext(r.Context(), "panic recovered",
				"panic", rec,
				"stack", string(debug.Stack()),
			)
			w.WriteHeader(http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}
