package http

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/handler/http/v1/dashboard"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Auth{CookieName: "opsybot_session"}
	h := dashboard.New(cfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, config.Ingest{})
	return NewRouter(slog.New(slog.DiscardHandler), cfg, config.Ingest{}, config.Telegram{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, h)
}

func TestRouterServesHealthPublicly(t *testing.T) {
	srv := httptest.NewServer(newTestRouter(t))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/v1/health")
	if err != nil {
		t.Fatalf("GET /v1/health: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200: health must be reachable without auth", res.StatusCode)
	}
}

func TestRouterRequiresAuthForProtectedRoute(t *testing.T) {
	srv := httptest.NewServer(newTestRouter(t))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/v1/me")
	if err != nil {
		t.Fatalf("GET /v1/me: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401: /me must require authentication", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}
