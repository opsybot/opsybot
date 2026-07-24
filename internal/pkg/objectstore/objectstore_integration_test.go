package objectstore

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/opsybot/opsybot/internal/config"
)

func testConfig(t *testing.T) config.S3 {
	t.Helper()
	endpoint := os.Getenv("OPSYBOT_TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("OPSYBOT_TEST_S3_ENDPOINT not set")
	}
	return config.S3{
		Endpoint:  endpoint,
		Region:    os.Getenv("OPSYBOT_TEST_S3_REGION"),
		Bucket:    os.Getenv("OPSYBOT_TEST_S3_BUCKET"),
		AccessKey: os.Getenv("OPSYBOT_TEST_S3_ACCESS_KEY"),
		SecretKey: os.Getenv("OPSYBOT_TEST_S3_SECRET_KEY"),
	}
}

func TestIntegrationObjectStoreRoundTrip(t *testing.T) {
	client, cleanup, err := New(testConfig(t))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(cleanup)
	if !client.Enabled() {
		t.Fatal("client reports disabled with a full configuration")
	}

	ctx := context.Background()
	key := "integration/objectstore-round-trip"
	payload := []byte("indexer queue depth 62144")

	if err := client.Put(ctx, key, bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("put: %v", err)
	}

	body, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got, err := io.ReadAll(body)
	body.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("read back %q, want %q", got, payload)
	}

	if err := client.Remove(ctx, key); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := client.Get(ctx, key); err == nil {
		t.Fatal("object still readable after Remove")
	}
}

func TestIntegrationObjectStoreDisabledWithoutEndpoint(t *testing.T) {
	client, cleanup, err := New(config.S3{Bucket: "opsybot", AccessKey: "k", SecretKey: "s"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(cleanup)
	if client.Enabled() {
		t.Fatal("client without an endpoint must report disabled")
	}

	ctx := context.Background()
	if err := client.Put(ctx, "k", bytes.NewReader(nil), 0, "text/plain"); err != ErrDisabled {
		t.Fatalf("put on a disabled client = %v, want ErrDisabled", err)
	}
	if _, err := client.Get(ctx, "k"); err != ErrDisabled {
		t.Fatalf("get on a disabled client = %v, want ErrDisabled", err)
	}
	if err := client.Remove(ctx, "k"); err != ErrDisabled {
		t.Fatalf("remove on a disabled client = %v, want ErrDisabled", err)
	}
}
