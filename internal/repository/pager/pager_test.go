package pager

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/webhook"
)

func TestDeliverToSignsPayloadAndReportsDelivered(t *testing.T) {
	payload := []byte(`{"event":"alert.notified","title":"db down"}`)
	var gotSig string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-Opsy-Signature")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	res, err := New(webhook.New(config.Webhook{})).DeliverTo(context.Background(), srv.URL, "topsecret", payload)
	if err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if !res.Delivered {
		t.Errorf("HTTP 202 should be delivered, got %+v", res)
	}
	if want := entity.SignBody("topsecret", gotBody); gotSig != want {
		t.Errorf("signature = %q, want %q", gotSig, want)
	}
}

func TestDeliverToOmitsSignatureWithoutSecret(t *testing.T) {
	got := "unset"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Opsy-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if _, err := New(webhook.New(config.Webhook{})).DeliverTo(context.Background(), srv.URL, "", []byte(`{}`)); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if got != "" {
		t.Errorf("no secret should send no signature header, got %q", got)
	}
}

func TestDeliverToUndeliveredOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	res, err := New(webhook.New(config.Webhook{})).DeliverTo(context.Background(), srv.URL, "", []byte(`{}`))
	if err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if res.Delivered {
		t.Errorf("HTTP 400 should not be delivered, got %+v", res)
	}
}
