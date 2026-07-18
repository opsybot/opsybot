package internal

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/secretbox"
	passwordreset "github.com/opsybot/opsybot/internal/repository/password_reset"
	"github.com/opsybot/opsybot/internal/repository/session"
	"github.com/opsybot/opsybot/internal/repository/user"
)

func TestIntegrationAuthFlow(t *testing.T) {
	dbURL := os.Getenv("OPSYBOT_TEST_POSTGRES_URL")
	if dbURL == "" {
		t.Skip("OPSYBOT_TEST_POSTGRES_URL not set")
	}

	client, cleanup, err := postgres.New(config.Postgres{
		URL:            dbURL,
		MaxOpenConns:   4,
		MaxIdleConns:   4,
		ConnectTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(cleanup)

	ctx := context.Background()
	migrator := &Migrator{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), PG: client}
	if err := migrator.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	box, err := secretbox.New(config.Auth{})
	if err != nil {
		t.Fatalf("secretbox: %v", err)
	}
	users := user.New(client, box)
	sessions := session.New(client)
	resets := passwordreset.New(client)

	stamp := time.Now().UnixNano()
	u, err := users.Create(ctx, entity.NewUser{
		Email:    fmt.Sprintf("authflow+%d@example.test", stamp),
		Name:     "Auth Flow",
		Timezone: "UTC",
	}, "argon2-not-verified-here")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = client.Querier(context.Background()).ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, u.ID)
	})

	t.Run("session lifecycle: create, resolve, expiry, revocation", func(t *testing.T) {
		now := time.Now()
		hash := fmt.Sprintf("sess-%d", stamp)
		sess, err := sessions.Create(ctx, u.ID, hash, "10.0.0.1", "it-agent", now.Add(time.Hour), now.Add(24*time.Hour))
		if err != nil {
			t.Fatalf("create session: %v", err)
		}

		got, err := sessions.GetByTokenHash(ctx, hash)
		if err != nil {
			t.Fatalf("get session: %v", err)
		}
		if got.ID != sess.ID || got.UserID != u.ID {
			t.Fatalf("session mismatch: got %+v", got)
		}
		if got.ExpiresAt.IsZero() || got.AbsoluteExpiresAt.IsZero() {
			t.Fatalf("expiry columns not persisted: %+v", got)
		}
		if !got.AbsoluteExpiresAt.After(got.ExpiresAt) {
			t.Fatalf("absolute expiry %v should be after idle expiry %v", got.AbsoluteExpiresAt, got.ExpiresAt)
		}

		if err := sessions.Delete(ctx, sess.ID); err != nil {
			t.Fatalf("revoke session: %v", err)
		}
		if _, err := sessions.GetByTokenHash(ctx, hash); err != entity.ErrSessionNotFound {
			t.Fatalf("after revoke, got err = %v, want ErrSessionNotFound", err)
		}
	})

	t.Run("revoke all sessions for a user", func(t *testing.T) {
		now := time.Now()
		for i := range 2 {
			if _, err := sessions.Create(ctx, u.ID, fmt.Sprintf("multi-%d-%d", stamp, i), "10.0.0.2", "it-agent",
				now.Add(time.Hour), now.Add(24*time.Hour)); err != nil {
				t.Fatalf("create session %d: %v", i, err)
			}
		}
		if err := sessions.DeleteByUser(ctx, u.ID); err != nil {
			t.Fatalf("delete by user: %v", err)
		}
		list, err := sessions.ListByUser(ctx, u.ID)
		if err != nil {
			t.Fatalf("list by user: %v", err)
		}
		if len(list) != 0 {
			t.Fatalf("after revoke-all, %d sessions remain", len(list))
		}
	})

	t.Run("password reset: usable, single-use, and expiry", func(t *testing.T) {
		now := time.Now()
		active := fmt.Sprintf("reset-active-%d", stamp)
		if err := resets.Create(ctx, u.ID, active, "10.0.0.3", now.Add(entity.PasswordResetTTL)); err != nil {
			t.Fatalf("create reset: %v", err)
		}
		r, err := resets.GetByTokenHash(ctx, active)
		if err != nil {
			t.Fatalf("get reset: %v", err)
		}
		if !r.Usable() {
			t.Fatalf("fresh reset should be usable: %+v", r)
		}
		if err := resets.MarkUsed(ctx, r.ID); err != nil {
			t.Fatalf("mark used: %v", err)
		}
		used, err := resets.GetByTokenHash(ctx, active)
		if err != nil {
			t.Fatalf("get used reset: %v", err)
		}
		if used.Usable() {
			t.Fatalf("used reset must not be usable: %+v", used)
		}

		expired := fmt.Sprintf("reset-expired-%d", stamp)
		if err := resets.Create(ctx, u.ID, expired, "10.0.0.4", now.Add(-time.Minute)); err != nil {
			t.Fatalf("create expired reset: %v", err)
		}
		e, err := resets.GetByTokenHash(ctx, expired)
		if err != nil {
			t.Fatalf("get expired reset: %v", err)
		}
		if e.Usable() {
			t.Fatalf("expired reset must not be usable: %+v", e)
		}
	})
}
