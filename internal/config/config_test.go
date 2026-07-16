package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

func TestNewDefaults(t *testing.T) {
	cfg, err := config.New("")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if cfg.Environment != "production" {
		t.Errorf("Environment = %q, want production", cfg.Environment)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 15s", cfg.ShutdownTimeout)
	}
	if cfg.Log.Level != "info" || cfg.Log.Format != "json" {
		t.Errorf("unexpected log defaults: %+v", cfg.Log)
	}
	p := cfg.Postgres
	if p.Host != "localhost" || p.Port != 5432 || p.User != "opsybot" ||
		p.Database != "opsybot" || p.SSLMode != "disable" {
		t.Errorf("unexpected postgres defaults: %+v", p)
	}
	if p.MaxOpenConns != 25 || p.MaxIdleConns != 25 {
		t.Errorf("unexpected pool defaults: %+v", p)
	}
	if p.ConnMaxLifetime != 5*time.Minute || p.ConnMaxIdleTime != 5*time.Minute ||
		p.ConnectTimeout != 5*time.Second {
		t.Errorf("unexpected duration defaults: %+v", p)
	}
}

func TestNewEnvOverridesDefaults(t *testing.T) {
	t.Setenv("OPSYBOT_POSTGRES_HOST", "db.internal")
	t.Setenv("OPSYBOT_POSTGRES_PORT", "6543")
	t.Setenv("OPSYBOT_POSTGRES_SSL_MODE", "require")
	t.Setenv("OPSYBOT_POSTGRES_CONNECT_TIMEOUT", "9s")
	t.Setenv("OPSYBOT_SHUTDOWN_TIMEOUT", "30s")

	cfg, err := config.New("")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if cfg.Postgres.Host != "db.internal" {
		t.Errorf("Host = %q, want db.internal", cfg.Postgres.Host)
	}
	if cfg.Postgres.Port != 6543 {
		t.Errorf("Port = %d, want 6543", cfg.Postgres.Port)
	}
	if cfg.Postgres.SSLMode != "require" {
		t.Errorf("SSLMode = %q, want require", cfg.Postgres.SSLMode)
	}
	if cfg.Postgres.ConnectTimeout != 9*time.Second {
		t.Errorf("ConnectTimeout = %v, want 9s", cfg.Postgres.ConnectTimeout)
	}
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 30s", cfg.ShutdownTimeout)
	}
}

func TestNewFileOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "opsybot.yaml")
	yaml := "environment: staging\npostgres:\n  host: filehost\n  database: filedb\n  conn_max_lifetime: 2m\n"
	if err := os.WriteFile(file, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.New(file)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if cfg.Environment != "staging" {
		t.Errorf("Environment = %q, want staging", cfg.Environment)
	}
	if cfg.Postgres.Host != "filehost" || cfg.Postgres.Database != "filedb" {
		t.Errorf("unexpected postgres from file: %+v", cfg.Postgres)
	}
	if cfg.Postgres.ConnMaxLifetime != 2*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 2m", cfg.Postgres.ConnMaxLifetime)
	}
	if cfg.Postgres.Port != 5432 {
		t.Errorf("Port = %d, want default 5432", cfg.Postgres.Port)
	}
}

func TestNewEnvBeatsFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "opsybot.yaml")
	if err := os.WriteFile(file, []byte("postgres:\n  host: filehost\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPSYBOT_POSTGRES_HOST", "envhost")

	cfg, err := config.New(file)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if cfg.Postgres.Host != "envhost" {
		t.Errorf("Host = %q, want envhost (env must beat file)", cfg.Postgres.Host)
	}
}

func TestNewMissingFileErrors(t *testing.T) {
	if _, err := config.New(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("New with missing file: want error, got nil")
	}
}

func TestNewPostgres(t *testing.T) {
	cfg := config.Config{Postgres: config.Postgres{Host: "h"}}
	if got := config.NewPostgres(cfg); got.Host != "h" {
		t.Fatalf("NewPostgres.Host = %q, want h", got.Host)
	}
}
