package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
)

func isolateEnv(t *testing.T) {
	t.Helper()
	for _, kv := range os.Environ() {
		key, value, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(key, "OPSYBOT_") {
			continue
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() { _ = os.Setenv(key, value) })
	}
}

func TestNewDefaults(t *testing.T) {
	isolateEnv(t)

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
	o := cfg.OTel
	if o.Endpoint != "" {
		t.Errorf("OTel.Endpoint = %q, want empty: tracing must be off unless configured", o.Endpoint)
	}
	if !o.Insecure || o.ServiceName != "opsybot" || o.SampleRatio != 1.0 {
		t.Errorf("unexpected otel defaults: %+v", o)
	}
	if o.MetricInterval != 60*time.Second || o.ExportTimeout != 10*time.Second {
		t.Errorf("unexpected otel duration defaults: %+v", o)
	}
	h := cfg.HTTP
	if h.Host != "" || h.Port != 8080 {
		t.Errorf("unexpected http defaults: %+v", h)
	}
	if h.ReadHeaderTimeout != 5*time.Second || h.ReadTimeout != 30*time.Second ||
		h.WriteTimeout != 30*time.Second || h.IdleTimeout != 120*time.Second {
		t.Errorf("unexpected http duration defaults: %+v", h)
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
	vk := cfg.Valkey
	if len(vk.Addrs) != 1 || vk.Addrs[0] != "localhost:6379" {
		t.Errorf("Valkey.Addrs = %v, want [localhost:6379]", vk.Addrs)
	}
	if vk.DB != 0 || vk.PoolSize != 10 {
		t.Errorf("unexpected valkey pool defaults: %+v", vk)
	}
	if vk.DialTimeout != 5*time.Second || vk.ReadTimeout != 3*time.Second ||
		vk.WriteTimeout != 3*time.Second {
		t.Errorf("unexpected valkey duration defaults: %+v", vk)
	}
	if cfg.Casbin.TableName != "casbin_rule" || cfg.Casbin.Channel != "/casbin" {
		t.Errorf("unexpected casbin defaults: %+v", cfg.Casbin)
	}
}

func TestNewEnvOverridesDefaults(t *testing.T) {
	isolateEnv(t)

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
	isolateEnv(t)

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
	isolateEnv(t)

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
	isolateEnv(t)

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

func TestNewValkeyAddrsFromCommaSeparatedEnv(t *testing.T) {
	isolateEnv(t)

	t.Setenv("OPSYBOT_VALKEY_ADDRS", "a:6379,b:6380")
	t.Setenv("OPSYBOT_VALKEY_DB", "3")

	cfg, err := config.New("")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got := cfg.Valkey.Addrs
	if len(got) != 2 || got[0] != "a:6379" || got[1] != "b:6380" {
		t.Errorf("Valkey.Addrs = %v, want [a:6379 b:6380]", got)
	}
	if cfg.Valkey.DB != 3 {
		t.Errorf("Valkey.DB = %d, want 3", cfg.Valkey.DB)
	}
}

func TestNewHTTP(t *testing.T) {
	cfg := config.Config{HTTP: config.HTTP{Port: 9999}}
	if got := config.NewHTTP(cfg); got.Port != 9999 {
		t.Fatalf("NewHTTP.Port = %d, want 9999", got.Port)
	}
}

func TestNewEnvironment(t *testing.T) {
	cfg := config.Config{Environment: "staging"}
	if got := config.NewEnvironment(cfg); got != config.Environment("staging") {
		t.Fatalf("NewEnvironment = %q, want staging", got)
	}
}

func TestNewOTel(t *testing.T) {
	cfg := config.Config{OTel: config.OTel{Endpoint: "collector:4317"}}
	if got := config.NewOTel(cfg); got.Endpoint != "collector:4317" {
		t.Fatalf("NewOTel.Endpoint = %q, want collector:4317", got.Endpoint)
	}
}

func TestNewCasbin(t *testing.T) {
	cfg := config.Config{Casbin: config.Casbin{TableName: "t"}}
	if got := config.NewCasbin(cfg); got.TableName != "t" {
		t.Fatalf("NewCasbin.TableName = %q, want t", got.TableName)
	}
}

func TestNewValkey(t *testing.T) {
	cfg := config.Config{Valkey: config.Valkey{Password: "p"}}
	if got := config.NewValkey(cfg); got.Password != "p" {
		t.Fatalf("NewValkey.Password = %q, want p", got.Password)
	}
}
