package ntfy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opsybot/opsybot/internal/config"
)

func TestPublishSendsMessageWithActions(t *testing.T) {
	var auth, agent, contentType string
	var body Message
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		auth = r.Header.Get("Authorization")
		agent = r.Header.Get("User-Agent")
		contentType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"id":"msg-1"}`))
	}))
	defer srv.Close()

	published, err := New(config.Ntfy{}).Publish(context.Background(), srv.URL, "tok", Message{
		Topic: "pages-x", Title: "Critical", Message: "db down", Priority: 5, Click: "https://x/alert",
		Actions: []Action{{Action: "http", Label: "Acknowledge", URL: "https://x/v1/act/atk", Method: "POST", Clear: true}},
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.ID != "msg-1" {
		t.Errorf("id = %q, want msg-1", published.ID)
	}
	if auth != "Bearer tok" {
		t.Errorf("auth = %q, want Bearer tok", auth)
	}
	if agent != "opsybot" {
		t.Errorf("user-agent = %q, want opsybot", agent)
	}
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("content-type = %q", contentType)
	}
	if body.Topic != "pages-x" || body.Priority != 5 {
		t.Errorf("body = %+v", body)
	}
	if len(body.Actions) != 1 || body.Actions[0].Label != "Acknowledge" || body.Actions[0].Method != "POST" {
		t.Errorf("actions = %+v", body.Actions)
	}
}

func TestPublishOmitsAuthWithoutToken(t *testing.T) {
	got := "unset"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"id":"m"}`))
	}))
	defer srv.Close()

	if _, err := New(config.Ntfy{}).Publish(context.Background(), srv.URL, "", Message{Topic: "t"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if got != "" {
		t.Errorf("no token should send no auth header, got %q", got)
	}
}

func TestPublishReturnsErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()

	_, err := New(config.Ntfy{}).Publish(context.Background(), srv.URL, "", Message{Topic: "t"})
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("want an error mentioning 500, got %v", err)
	}
}

func TestPublishFallsBackToDefaultServer(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hit = true
		_, _ = w.Write([]byte(`{"id":"m"}`))
	}))
	defer srv.Close()

	if _, err := New(config.Ntfy{DefaultServer: srv.URL}).Publish(context.Background(), "", "", Message{Topic: "t"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if !hit {
		t.Fatal("an empty per-message server should fall back to the configured default server")
	}
}
