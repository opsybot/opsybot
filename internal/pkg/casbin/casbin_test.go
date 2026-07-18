package casbin

import (
	"testing"

	casbinv3 "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/pkg/valkey"
)

type countingAdapter struct{ loads int }

func (a *countingAdapter) LoadPolicy(m model.Model) error                      { a.loads++; return nil }
func (a *countingAdapter) SavePolicy(m model.Model) error                      { return nil }
func (a *countingAdapter) AddPolicy(sec, ptype string, rule []string) error    { return nil }
func (a *countingAdapter) RemovePolicy(sec, ptype string, rule []string) error { return nil }
func (a *countingAdapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

func newEnforcer(t *testing.T) *casbinv3.SyncedEnforcer {
	t.Helper()
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		t.Fatalf("NewModelFromString: %v", err)
	}
	e, err := casbinv3.NewSyncedEnforcer(m)
	if err != nil {
		t.Fatalf("NewSyncedEnforcer: %v", err)
	}
	return e
}

func TestModelScopesRolesToTheirDomain(t *testing.T) {
	e := newEnforcer(t)
	if _, err := e.AddPolicy("admin", "ws-a", "incidents", "read"); err != nil {
		t.Fatalf("AddPolicy: %v", err)
	}
	if _, err := e.AddGroupingPolicy("alice", "admin", "ws-a"); err != nil {
		t.Fatalf("AddGroupingPolicy: %v", err)
	}

	tests := []struct {
		name string
		sub  string
		dom  string
		obj  string
		act  string
		want bool
	}{
		{"role grants in its own domain", "alice", "ws-a", "incidents", "read", true},
		{"role does not leak across domains", "alice", "ws-b", "incidents", "read", false},
		{"role does not grant another action", "alice", "ws-a", "incidents", "write", false},
		{"unknown subject is denied", "mallory", "ws-a", "incidents", "read", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := e.Enforce(tc.sub, tc.dom, tc.obj, tc.act)
			if err != nil {
				t.Fatalf("Enforce: %v", err)
			}
			if got != tc.want {
				t.Errorf("Enforce(%q, %q, %q, %q) = %v, want %v", tc.sub, tc.dom, tc.obj, tc.act, got, tc.want)
			}
		})
	}
}

func TestWatcherOptionsReloadPolicyOnUpdate(t *testing.T) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		t.Fatalf("NewModelFromString: %v", err)
	}
	a := &countingAdapter{}
	e, err := casbinv3.NewSyncedEnforcer(m, a)
	if err != nil {
		t.Fatalf("NewSyncedEnforcer: %v", err)
	}
	loadsAfterInit := a.loads

	opts := watcherOptions(e, config.Casbin{Channel: "/casbin"}, valkey.Client{})
	if opts.OptionalUpdateCallback == nil {
		t.Fatal("OptionalUpdateCallback is nil: SetWatcher does not wire a callback for a WatcherEx, so policy would never reload")
	}

	opts.OptionalUpdateCallback("")
	if a.loads != loadsAfterInit+1 {
		t.Errorf("adapter loads = %d, want %d: update callback must reload policy", a.loads, loadsAfterInit+1)
	}
}

func TestWatcherOptionsIgnoreSelfAndChannel(t *testing.T) {
	e := newEnforcer(t)
	opts := watcherOptions(e, config.Casbin{Channel: "/opsybot-casbin"}, valkey.Client{})

	if !opts.IgnoreSelf {
		t.Error("IgnoreSelf = false, want true: an instance must not reload from its own update")
	}
	if opts.Channel != "/opsybot-casbin" {
		t.Errorf("Channel = %q, want /opsybot-casbin", opts.Channel)
	}
	if opts.SubClient == nil || opts.PubClient == nil {
		t.Error("Sub/PubClient must be set, otherwise the watcher dials its own connection")
	}
}
