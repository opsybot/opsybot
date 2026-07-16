package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

func TestIntegrationPostgres(t *testing.T) {
	dbURL := os.Getenv("OPSYBOT_TEST_POSTGRES_URL")
	if dbURL == "" {
		t.Skip("OPSYBOT_TEST_POSTGRES_URL not set")
	}

	c, cleanup, err := New(config.Postgres{
		URL:            dbURL,
		MaxOpenConns:   2,
		MaxIdleConns:   2,
		ConnectTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(cleanup)

	ctx := context.Background()
	if _, err := c.Querier(ctx).ExecContext(ctx,
		`CREATE TEMPORARY TABLE it_probe (id int PRIMARY KEY)`); err != nil {
		t.Fatalf("create temp table: %v", err)
	}

	err = c.WithTx(ctx, func(ctx context.Context) error {
		_, err := c.Querier(ctx).ExecContext(ctx, `INSERT INTO it_probe (id) VALUES (1)`)
		return err
	})
	if err != nil {
		t.Fatalf("WithTx commit: %v", err)
	}

	sentinel := errors.New("force rollback")
	err = c.WithTx(ctx, func(ctx context.Context) error {
		if _, err := c.Querier(ctx).ExecContext(ctx, `INSERT INTO it_probe (id) VALUES (2)`); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTx rollback: err = %v, want %v", err, sentinel)
	}

	var n int
	if err := c.Querier(ctx).QueryRowContext(ctx,
		`SELECT count(*) FROM it_probe`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Fatalf("row count = %d, want 1", n)
	}
}
