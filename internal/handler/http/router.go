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
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/handler/http/middleware"
	"github.com/opsybot/opsybot/internal/service"
	dashboardapi "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

const (
	dashboardBaseURL = "/v1"
	tracerName       = "opsybot/http"
)

func NewRouter(log *slog.Logger, cfg config.Auth, cfgIngest config.Ingest, cfgTelegram config.Telegram, auth service.Auth, keys service.APIKeys, sso service.SSO, schedules service.Schedules, ingest service.Ingest, limiter service.RateLimiter, channels service.Channels, interactions service.Interactions, actions service.Actions, chats service.Chats, dashboard dashboardapi.StrictServerInterface) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(otelMiddleware)
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.ClientInfo(cfg.TrustProxyHeaders))
	r.Use(middleware.RateLimit(limiter))
	r.Use(middleware.Authn(auth, keys, cfg))

	ssoRoutes := &ssoRoutes{sso: sso, cfg: cfg}
	r.Get("/v1/auth/sso/{workspace}/start", ssoRoutes.start)
	r.Get("/v1/auth/sso/{workspace}/callback", ssoRoutes.callback)
	r.Get("/v1/auth/sso/{workspace}/saml/metadata", ssoRoutes.samlMetadata)
	r.Post("/v1/auth/sso/{workspace}/saml/acs", ssoRoutes.samlACS)

	feedRoutes := &feedRoutes{schedules: schedules}
	r.Get("/v1/oncall/feed/{token}", feedRoutes.serve)

	ingestRoutes := newIngestRoutes(ingest, cfgIngest)
	r.Post("/v1/ingest/e/{token}", ingestRoutes.webhook)
	r.Get("/v1/ingest/hb/{token}", ingestRoutes.checkIn)
	r.Post("/v1/ingest/hb/{token}", ingestRoutes.checkIn)

	verifyRoutes := &channelVerifyRoutes{channels: channels}
	r.Get("/v1/channels/verify/{token}", verifyRoutes.confirm)
	r.Post("/v1/channels/verify/{token}", verifyRoutes.confirm)

	chatRoutes := &chatCallbackRoutes{interactions: interactions}
	r.Post("/v1/chat/slack/interactions", chatRoutes.slack)
	r.Post("/v1/chat/discord/interactions", chatRoutes.discord)

	chatOAuthRoutes := &chatOAuthRoutes{chats: chats, auth: auth, cfg: cfg}
	r.Get("/v1/chat/slack/oauth/callback", chatOAuthRoutes.callback(entity.ChatProviderSlack))
	r.Get("/v1/chat/slack/identity/callback", chatOAuthRoutes.identityCallback(entity.ChatProviderSlack))
	r.Get("/v1/chat/discord/oauth/callback", chatOAuthRoutes.callback(entity.ChatProviderDiscord))
	r.Get("/v1/chat/discord/identity/callback", chatOAuthRoutes.identityCallback(entity.ChatProviderDiscord))

	telegramRoutes := &telegramWebhookRoutes{chats: chats, actions: actions, cfg: cfgTelegram}
	r.Post("/v1/chat/telegram/hook/{secret}", telegramRoutes.handle)

	actionRoutes := &actionLinkRoutes{actions: actions}
	r.Get("/v1/act/{token}", actionRoutes.prompt)
	r.Post("/v1/act/{token}", actionRoutes.act)

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
