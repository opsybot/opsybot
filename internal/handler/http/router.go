package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/handler/http/middleware"
	"github.com/opsybot/opsybot/internal/service"
	dashboardapi "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

const (
	dashboardBaseURL = "/v1"
	tracerName       = "opsybot/http"
)

func NewRouter(log *slog.Logger, cfg config.Auth, auth service.Auth, keys service.APIKeys, sso service.SSO, dashboard dashboardapi.StrictServerInterface) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(otelMiddleware)
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.ClientInfo(cfg.TrustProxyHeaders))
	r.Use(middleware.Authn(auth, keys, cfg))

	ssoRoutes := &ssoRoutes{sso: sso, cfg: cfg}
	r.Get("/v1/auth/sso/{workspace}/start", ssoRoutes.start)
	r.Get("/v1/auth/sso/{workspace}/callback", ssoRoutes.callback)

	dashboardapi.HandlerWithOptions(
		dashboardapi.NewStrictHandler(dashboard, nil),
		dashboardapi.ChiServerOptions{
			BaseURL:    dashboardBaseURL,
			BaseRouter: r,
		},
	)

	return r
}

func otelMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		pattern := chi.RouteContext(r.Context()).RoutePattern()
		if pattern == "" {
			return
		}
		span := trace.SpanFromContext(r.Context())
		span.SetName(r.Method + " " + pattern)
		span.SetAttributes(semconv.HTTPRoute(pattern))
	}), tracerName)
}
