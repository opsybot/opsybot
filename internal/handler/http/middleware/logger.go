package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"

	"github.com/opsybot/opsybot/internal/pkg/logger"
)

func Logger(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqLog := l.With(
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
			)
			if sc := trace.SpanContextFromContext(r.Context()); sc.IsValid() {
				reqLog = reqLog.With("trace_id", sc.TraceID().String())
			}
			ctx := logger.Into(r.Context(), reqLog)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			next.ServeHTTP(ww, r.WithContext(ctx))

			reqLog.InfoContext(ctx, "request handled",
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start),
			)
		})
	}
}
