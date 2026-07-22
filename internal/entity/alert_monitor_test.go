package entity

import (
	"testing"
	"time"
)

func monitor() AlertMonitor {
	return AlertMonitor{
		ID:        "mon-1",
		Slug:      "nightly-backup",
		Name:      "nightly-backup",
		Interval:  time.Hour,
		Grace:     10 * time.Minute,
		Severity:  SeverityHigh,
		CreatedAt: time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC),
	}
}

func TestMonitorStateFlipsAfterIntervalPlusGrace(t *testing.T) {
	m := monitor()
	m.LastCheckInAt = time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC)

	healthy := m.LastCheckInAt.Add(time.Hour + 9*time.Minute)
	if got := m.State(healthy); got != MonitorStateHealthy {
		t.Errorf("state at interval+9m = %q, want healthy: grace has not run out yet", got)
	}

	missed := m.LastCheckInAt.Add(time.Hour + 11*time.Minute)
	if got := m.State(missed); got != MonitorStateMissed {
		t.Errorf("state at interval+11m = %q, want missed", got)
	}
}

func TestMonitorPausedOverridesMissed(t *testing.T) {
	m := monitor()
	m.Paused = true
	m.LastCheckInAt = time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)

	if got := m.State(time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)); got != MonitorStatePaused {
		t.Errorf("state = %q, want paused: a paused monitor must not read as missed", got)
	}
}

func TestMonitorWithoutCheckInCountsFromCreation(t *testing.T) {
	m := monitor()

	if want := m.CreatedAt.Add(time.Hour + 10*time.Minute); !m.DueAt().Equal(want) {
		t.Errorf("DueAt() = %v, want %v: a monitor that never checked in is due from creation", m.DueAt(), want)
	}
	if got := m.State(m.CreatedAt.Add(2 * time.Hour)); got != MonitorStateMissed {
		t.Errorf("state = %q, want missed: a job that never checked in must page", got)
	}
}

func TestMonitorDedupKeyIsNamespacedPerMonitor(t *testing.T) {
	a := monitor()
	b := monitor()
	b.ID = "mon-2"

	if a.DedupKey() == b.DedupKey() {
		t.Fatal("two monitors share a dedup key, so one missed check-in would swallow the other")
	}
	if got := a.DedupKey(); got != MonitorDedupPrefix+"mon-1" {
		t.Errorf("DedupKey() = %q, want the reserved monitor namespace", got)
	}
}

func TestNewAlertMonitorRejectsOutOfRangeInterval(t *testing.T) {
	in := NewAlertMonitor{
		Name:      "nightly-backup",
		Slug:      "nightly-backup",
		Interval:  30 * time.Second,
		Grace:     time.Minute,
		PolicyRef: DefaultPolicyRef,
		Severity:  SeverityHigh,
	}
	if err := in.Validate(); !IsValidationError(err) {
		t.Fatalf("Validate() = %v, want a validation error for a sub-minute interval", err)
	}

	in.Interval = MonitorIntervalMax + time.Hour
	if err := in.Validate(); !IsValidationError(err) {
		t.Fatalf("Validate() = %v, want a validation error for an interval past the maximum", err)
	}

	in.Interval = time.Hour
	if err := in.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil for an hourly monitor", err)
	}
}
