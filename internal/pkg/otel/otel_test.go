package otel

import (
	"context"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

func TestNewWithoutEndpointIsNoop(t *testing.T) {
	c, cleanup, err := New(config.OTel{ServiceName: "opsybot"}, "test")
	if err != nil {
		t.Fatalf("New with no endpoint must not fail: %v", err)
	}
	if cleanup == nil {
		t.Fatal("cleanup must never be nil")
	}
	if c.tracer != nil || c.meter != nil {
		t.Error("no providers may be built without an endpoint")
	}
	cleanup()
}

func TestNewWithUnreachableEndpointStillStarts(t *testing.T) {
	c, cleanup, err := New(config.OTel{
		Endpoint:       "127.0.0.1:1",
		Insecure:       true,
		ServiceName:    "opsybot",
		SampleRatio:    1,
		MetricInterval: time.Minute,
		ExportTimeout:  time.Second,
	}, "test")
	if err != nil {
		t.Fatalf("a down collector must not stop the app from starting: %v", err)
	}
	t.Cleanup(cleanup)
	if c.tracer == nil || c.meter == nil {
		t.Fatal("providers must be built when an endpoint is configured")
	}
}

func TestFromReturnsSpanFromContext(t *testing.T) {
	if span := From(context.Background()); span == nil {
		t.Fatal("From must always return a span, never nil")
	}
	if From(context.Background()).SpanContext().IsValid() {
		t.Error("a background context must not carry a valid span context")
	}
}
