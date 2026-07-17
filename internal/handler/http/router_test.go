package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsybot/opsybot/internal/handler/http/v1/dashboard"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	return NewRouter(slog.New(slog.DiscardHandler), dashboard.New())
}

func TestRouterServesHealthUnderBaseURL(t *testing.T) {
	srv := httptest.NewServer(newTestRouter(t))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/v1/health")
	if err != nil {
		t.Fatalf("GET /v1/health: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var health struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("decode %q: %v", body, err)
	}
	if health.Status != "ok" {
		t.Errorf("status = %q, want ok", health.Status)
	}
}

func TestRouterDoesNotServeHealthAtRoot(t *testing.T) {
	srv := httptest.NewServer(newTestRouter(t))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404: routes must be served under %s only", res.StatusCode, dashboardBaseURL)
	}
}

func TestRouterRecoversFromPanic(t *testing.T) {
	r := newTestRouter(t).(interface {
		http.Handler
		Get(string, http.HandlerFunc)
	})
	r.Get("/boom", func(http.ResponseWriter, *http.Request) { panic("boom") })

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/boom")
	if err != nil {
		t.Fatalf("GET /boom: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500: a panic must not kill the server", res.StatusCode)
	}
}
