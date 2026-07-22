package entity

import (
	"errors"
	"strings"
	"time"
)

type IngestOutcome string

const (
	IngestOutcomeCreated      IngestOutcome = "created"
	IngestOutcomeUpdated      IngestOutcome = "updated"
	IngestOutcomeDuplicate    IngestOutcome = "duplicate"
	IngestOutcomeResolved     IngestOutcome = "resolved"
	IngestOutcomeStale        IngestOutcome = "stale"
	IngestOutcomeFailed       IngestOutcome = "failed"
	IngestOutcomeFloodDropped IngestOutcome = "flood_dropped"
)

type IngestFailureReason string

const (
	FailureInvalidJSON        IngestFailureReason = "invalid_json"
	FailureEmptyBody          IngestFailureReason = "empty_body"
	FailureBodyTooLarge       IngestFailureReason = "body_too_large"
	FailureUnsupportedFormat  IngestFailureReason = "unsupported_format"
	FailureUnsupportedVersion IngestFailureReason = "unsupported_version"
	FailureMissingTitle       IngestFailureReason = "missing_title"
	FailureNoAlerts           IngestFailureReason = "no_alerts"
	FailureBadTimestamp       IngestFailureReason = "bad_timestamp"
	FailureSignatureInvalid   IngestFailureReason = "signature_invalid"
	FailureMappingUnresolved  IngestFailureReason = "mapping_unresolved"
)

var (
	ErrIngestBodyEmpty    = errors.New("ingest body empty")
	ErrIngestBodyTooLarge = errors.New("ingest body too large")
	ErrIngestUnparseable  = errors.New("ingest body unparseable")
	ErrIngestFlooded      = errors.New("ingest source flooded")
)

type IngestParseError struct {
	Reason IngestFailureReason
	Detail string
}

func (e IngestParseError) Error() string {
	return string(e.Reason) + ": " + e.Detail
}

func (e IngestParseError) Unwrap() error {
	return ErrIngestUnparseable
}

func ParseFailure(reason IngestFailureReason, detail string) error {
	return IngestParseError{Reason: reason, Detail: detail}
}

func ParseFailureOf(err error) (IngestParseError, bool) {
	return errors.AsType[IngestParseError](err)
}

type IngestRequest struct {
	Token       string
	Signature   string
	Body        []byte
	ContentType string
	Method      string
	RemoteIP    string
	ReceivedAt  time.Time
}

type IngestedAlert struct {
	DedupKeyRaw string
	Title       string
	Description string
	Severity    AlertSeverity
	SourceLabel string
	ServiceName string
	Labels      map[string]string
	Links       []AlertLink
	StartedAt   time.Time
	EndedAt     time.Time
	Resolved    bool
	ResolveMode ResolveMode
	Repeat      bool
	Payload     string
}

type IngestResult struct {
	AlertID  string
	DedupKey string
	Outcome  IngestOutcome
}

type IngestFailure struct {
	ID          string
	WorkspaceID string
	SourceID    string
	SourceSlug  string
	Reason      IngestFailureReason
	Detail      string
	Payload     string
	At          time.Time
}

type IngestEvent struct {
	ID          string
	WorkspaceID string
	SourceID    string
	AlertID     string
	DedupKey    string
	Outcome     IngestOutcome
	At          time.Time
}

func (a IngestedAlert) Normalize(src AlertSource, now time.Time) IngestedAlert {
	out := a
	out.Title = truncate(strings.TrimSpace(out.Title), AlertTitleMaxLength)
	out.Description = truncate(out.Description, AlertDescriptionMaxLength)
	out.Payload = truncate(out.Payload, AlertPayloadMaxBytes)
	out.Labels = TruncateLabels(out.Labels)

	if out.Severity == "" {
		out.Severity = src.DefaultSeverity
	}
	if out.Severity == "" {
		out.Severity = SeverityWarning
	}
	if out.SourceLabel == "" {
		out.SourceLabel = src.Slug
	}
	if out.StartedAt.IsZero() {
		out.StartedAt = now
	}
	if out.Resolved && out.EndedAt.IsZero() {
		out.EndedAt = now
	}
	if out.Resolved && out.ResolveMode == "" {
		out.ResolveMode = ResolveModeSource
	}
	if len(out.Links) > AlertLinksMax {
		out.Links = out.Links[:AlertLinksMax]
	}
	return out
}

func (a IngestedAlert) Valid() bool {
	return strings.TrimSpace(a.Title) != ""
}
