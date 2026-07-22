package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
)

type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityHigh     AlertSeverity = "high"
	SeverityWarning  AlertSeverity = "warning"
)

type AlertStatus string

const (
	AlertStatusOpen     AlertStatus = "open"
	AlertStatusAcked    AlertStatus = "acked"
	AlertStatusResolved AlertStatus = "resolved"
)

type ResolveMode string

const (
	ResolveModeSource   ResolveMode = "source"
	ResolveModeManual   ResolveMode = "manual"
	ResolveModeIncident ResolveMode = "incident"
	ResolveModeTimeout  ResolveMode = "timeout"
)

type AlertLinkKind string

const (
	AlertLinkRunbook   AlertLinkKind = "runbook"
	AlertLinkDashboard AlertLinkKind = "dashboard"
	AlertLinkSource    AlertLinkKind = "source"
)

type AlertEventKind string

const (
	AlertEventReceived   AlertEventKind = "received"
	AlertEventDeduped    AlertEventKind = "deduped"
	AlertEventGrouped    AlertEventKind = "grouped"
	AlertEventRouted     AlertEventKind = "routed"
	AlertEventSuppressed AlertEventKind = "suppressed"
	AlertEventAcked      AlertEventKind = "acked"
	AlertEventResolved   AlertEventKind = "resolved"
)

const (
	AlertTitleMaxLength       = 255
	AlertDescriptionMaxLength = 15000
	AlertLabelsMaxEntries     = 64
	AlertLabelsMaxBytes       = 8192
	AlertLabelKeyMaxLength    = 100
	AlertLabelValueMaxLength  = 400
	AlertLinksMax             = 10
	AlertPayloadMaxBytes      = 65536
	DedupKeyMaxLength         = 512
	AlertTimelineLimit        = 200
	AlertListMaxPageSize      = 100
	AlertListDefaultPageSize  = 50
	FloodDedupPrefix          = "opsybot/flood/"
	MonitorDedupPrefix        = "opsybot/monitor/"
	GroupDedupPrefix          = "opsybot/group/"
)

type AlertLink struct {
	Kind  AlertLinkKind
	Label string
	URL   string
}

type AlertChild struct {
	ID         string
	Title      string
	Count      int
	LastSeenAt time.Time
}

type AlertEvent struct {
	ID      string
	AlertID string
	At      time.Time
	Kind    AlertEventKind
	Text    string
	Result  string
}

type Alert struct {
	ID                    string
	WorkspaceID           string
	SourceID              string
	SourceSlug            string
	ParentAlertID         string
	DedupKey              string
	GroupKey              string
	Title                 string
	Description           string
	Severity              AlertSeverity
	Status                AlertStatus
	SourceLabel           string
	ServiceName           string
	Labels                map[string]string
	Count                 int
	StartedAt             time.Time
	LastSeenAt            time.Time
	EndedAt               time.Time
	AckedAt               time.Time
	ResolvedAt            time.Time
	AckedByUserID         string
	AckedByLabel          string
	ResolveMode           ResolveMode
	RoutedPolicyRef       string
	SuppressedBySilenceID string
	SuppressedAt          time.Time
	Payload               string
	Links                 []AlertLink
	Children              []AlertChild
	Timeline              []AlertEvent
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type AlertUpsert struct {
	WorkspaceID string
	SourceID    string
	DedupKey    string
	Title       string
	Description string
	Severity    AlertSeverity
	SourceLabel string
	ServiceName string
	Labels      map[string]string
	StartedAt   time.Time
	LastSeenAt  time.Time
	Payload     string
	Links       []AlertLink
}

type AlertFilter struct {
	Statuses   []AlertStatus
	Severities []AlertSeverity
	SourceIDs  []string
	Query      string
	Cursor     string
	Limit      int
}

var (
	ErrAlertNotFound        = errors.New("alert not found")
	ErrAlertAlreadyResolved = errors.New("alert already resolved")
	ErrAlertBulkEmpty       = errors.New("alert bulk selection empty")
)

func (s AlertSeverity) Validate() error {
	return alertSeverityField(s)
}

func (s AlertStatus) Validate() error {
	return alertStatusField(s)
}

func NormalizeSeverity(raw string, fallback AlertSeverity) AlertSeverity {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "critical", "crit", "fatal", "emergency", "page", "p1", "sev1":
		return SeverityCritical
	case "high", "error", "err", "major", "p2", "sev2":
		return SeverityHigh
	case "warning", "warn", "minor", "p3", "sev3":
		return SeverityWarning
	case "info", "information", "informational", "low", "debug", "p4", "p5", "sev4":
		return SeverityWarning
	default:
		if fallback == "" {
			return SeverityWarning
		}
		return fallback
	}
}

func CanonicalLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(labels[k])
	}
	return b.String()
}

func DeriveDedupKey(sourceID, raw, title, sourceLabel string, labels map[string]string) string {
	if trimmed := strings.TrimSpace(raw); trimmed != "" {
		if len(trimmed) > DedupKeyMaxLength {
			sum := sha256.Sum256([]byte(trimmed))
			return hex.EncodeToString(sum[:])
		}
		return trimmed
	}
	sum := sha256.Sum256([]byte(strings.Join([]string{sourceID, title, sourceLabel, CanonicalLabels(labels)}, "\x00")))
	return hex.EncodeToString(sum[:])
}

func TruncateLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return map[string]string{}
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(map[string]string, len(keys))
	bytes := 0
	for _, k := range keys {
		if len(out) >= AlertLabelsMaxEntries {
			break
		}
		key := truncate(k, AlertLabelKeyMaxLength)
		value := truncate(labels[k], AlertLabelValueMaxLength)
		if bytes+len(key)+len(value) > AlertLabelsMaxBytes {
			break
		}
		bytes += len(key) + len(value)
		out[key] = value
	}
	return out
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func (a Alert) Open() bool {
	return a.Status != AlertStatusResolved
}
