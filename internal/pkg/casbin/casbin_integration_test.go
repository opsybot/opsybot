package casbin

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/valkey"
)

func dropTableOnCleanup(t *testing.T, pgCfg config.Postgres, table string) {
	t.Helper()
	t.Cleanup(func() {
		pg, cleanup, err := postgres.New(pgCfg)
		if err != nil {
			t.Errorf("connect to drop %s: %v", table, err)
			return
		}
		defer cleanup()
		if _, err := pg.ExecContext(context.Background(), "DROP TABLE IF EXISTS "+table); err != nil {
			t.Errorf("drop %s: %v", table, err)
		}
	})
}

func testConfig(t *testing.T) (config.Postgres, config.Valkey, config.Casbin) {
	t.Helper()
	dbURL := os.Getenv("OPSYBOT_TEST_POSTGRES_URL")
	if dbURL == "" {
		t.Skip("OPSYBOT_TEST_POSTGRES_URL not set")
	}
	addrs := os.Getenv("OPSYBOT_TEST_VALKEY_ADDRS")
	if addrs == "" {
		t.Skip("OPSYBOT_TEST_VALKEY_ADDRS not set")
	}
	return config.Postgres{
			URL:            dbURL,
			MaxOpenConns:   4,
			MaxIdleConns:   4,
			ConnectTimeout: 5 * time.Second,
		}, config.Valkey{
			Addrs:       strings.Split(addrs, ","),
			DialTimeout: 5 * time.Second,
			PoolSize:    4,
		}, config.Casbin{
			TableName: "casbin_rule_integration_test",
			Channel:   "/casbin-integration-test",
		}
}

func newInstance(t *testing.T, pgCfg config.Postgres, vkCfg config.Valkey, cbCfg config.Casbin) Client {
	t.Helper()
	pg, pgCleanup, err := postgres.New(pgCfg)
	if err != nil {
		t.Fatalf("postgres.New: %v", err)
	}
	t.Cleanup(pgCleanup)

	vk, vkCleanup, err := valkey.New(vkCfg)
	if err != nil {
		t.Fatalf("valkey.New: %v", err)
	}
	t.Cleanup(vkCleanup)

	c, cleanup, err := New(cbCfg, pg, vk)
	if err != nil {
		t.Fatalf("casbin.New: %v", err)
	}
	t.Cleanup(cleanup)
	return c
}

func TestIntegrationPolicyChangePropagatesToOtherInstance(t *testing.T) {
	pgCfg, vkCfg, cbCfg := testConfig(t)
	dropTableOnCleanup(t, pgCfg, cbCfg.TableName)

	first := newInstance(t, pgCfg, vkCfg, cbCfg)
	second := newInstance(t, pgCfg, vkCfg, cbCfg)

	policy := []any{"admin", "ws-propagation", "incidents", "read"}
	if ok, err := first.AddPolicy(policy...); err != nil {
		t.Fatalf("AddPolicy: %v", err)
	} else if !ok {
		t.Fatal("AddPolicy returned false, policy already present")
	}
	t.Cleanup(func() { _, _ = first.RemovePolicy(policy...) })

	deadline := time.Now().Add(10 * time.Second)
	for {
		ok, err := second.HasPolicy(policy...)
		if err != nil {
			t.Fatalf("HasPolicy: %v", err)
		}
		if ok {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("second instance never saw the policy: the watcher update callback did not reload policy")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
