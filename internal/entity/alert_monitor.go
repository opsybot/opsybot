package entity

import (
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type MonitorState string

const (
	MonitorStateHealthy MonitorState = "healthy"
	MonitorStateMissed  MonitorState = "missed"
	MonitorStatePaused  MonitorState = "paused"
)

const (
	MonitorIntervalMin     = time.Minute
	MonitorIntervalMax     = 30 * 24 * time.Hour
	MonitorIntervalDefault = time.Hour
	MonitorGraceMax        = 24 * time.Hour
	MonitorGraceDefault    = 5 * time.Minute
	MonitorSweepBatch      = 200
)

type AlertMonitor struct {
	ID            string
	WorkspaceID   string
	SourceID      string
	Slug          string
	Name          string
	Interval      time.Duration
	Grace         time.Duration
	PolicyID      string
	PolicySlug    string
	Severity      AlertSeverity
	LastCheckInAt time.Time
	Paused        bool
	CheckInToken  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NewAlertMonitor struct {
	Name       string
	Slug       string
	Interval   time.Duration
	Grace      time.Duration
	PolicySlug string
	PolicyID   string
	Severity   AlertSeverity
}

type AlertMonitorUpdate struct {
	Name       string
	Interval   time.Duration
	Grace      time.Duration
	PolicySlug string
	PolicyID   string
	Severity   AlertSeverity
}

var (
	ErrAlertMonitorNotFound = errors.New("alert monitor not found")
	ErrAlertMonitorFormat   = errors.New("alert source is not a heartbeat monitor")
)

func (s MonitorState) Validate() error {
	return monitorStateField(s)
}

func (m AlertMonitor) Since() time.Time {
	if !m.LastCheckInAt.IsZero() {
		return m.LastCheckInAt
	}
	return m.CreatedAt
}

func (m AlertMonitor) DueAt() time.Time {
	return m.Since().Add(m.Interval + m.Grace)
}

func (m AlertMonitor) State(now time.Time) MonitorState {
	switch {
	case m.Paused:
		return MonitorStatePaused
	case now.After(m.DueAt()):
		return MonitorStateMissed
	default:
		return MonitorStateHealthy
	}
}

func (m AlertMonitor) DedupKey() string {
	return MonitorDedupPrefix + m.ID
}

func (m AlertMonitor) MissedTitle() string {
	return m.Name + " missed its check-in"
}

func (m AlertMonitor) MissedDescription(now time.Time) string {
	return "Expected a check-in every " + FormatDuration(m.Interval) + " with " + FormatDuration(m.Grace) +
		" of grace. Nothing has arrived for " + FormatDuration(now.Sub(m.Since())) + "."
}

func (n NewAlertMonitor) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Name, validation.By(sourceNameField)),
		validation.Field(&n.Slug, validation.By(sourceSlugField)),
		validation.Field(&n.Interval, validation.By(monitorIntervalField)),
		validation.Field(&n.Grace, validation.By(monitorGraceField)),
		validation.Field(&n.PolicySlug, validation.By(policyRefField)),
		validation.Field(&n.Severity, validation.By(alertSeverityField)),
	)
}

func (u AlertMonitorUpdate) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.By(sourceNameField)),
		validation.Field(&u.Interval, validation.By(monitorIntervalField)),
		validation.Field(&u.Grace, validation.By(monitorGraceField)),
		validation.Field(&u.PolicySlug, validation.By(policyRefField)),
		validation.Field(&u.Severity, validation.By(alertSeverityField)),
	)
}
