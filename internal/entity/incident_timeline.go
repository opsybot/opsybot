package entity

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type IncidentEventKind string

const (
	IncidentEventDeclared        IncidentEventKind = "declared"
	IncidentEventStatusChanged   IncidentEventKind = "status_changed"
	IncidentEventSeverityChanged IncidentEventKind = "severity_changed"
	IncidentEventLeadChanged     IncidentEventKind = "lead_changed"
	IncidentEventRenamed         IncidentEventKind = "renamed"
	IncidentEventSummaryChanged  IncidentEventKind = "summary_changed"
	IncidentEventFieldsChanged   IncidentEventKind = "fields_changed"
	IncidentEventReopened        IncidentEventKind = "reopened"
	IncidentEventResolved        IncidentEventKind = "resolved"
	IncidentEventAlertLinked     IncidentEventKind = "alert_linked"
	IncidentEventAlertUnlinked   IncidentEventKind = "alert_unlinked"
	IncidentEventRelated         IncidentEventKind = "related"
	IncidentEventUnrelated       IncidentEventKind = "unrelated"
	IncidentEventFollowupAdded   IncidentEventKind = "followup_added"
	IncidentEventFollowupDone    IncidentEventKind = "followup_done"
	IncidentEventUpdated         IncidentEventKind = "updated"
	IncidentEventNote            IncidentEventKind = "note"
)

type IncidentEventCategory string

const (
	IncidentCategoryStatus        IncidentEventCategory = "status"
	IncidentCategoryCommunication IncidentEventCategory = "communication"
	IncidentCategoryAction        IncidentEventCategory = "action"
	IncidentCategoryObservation   IncidentEventCategory = "observation"
	IncidentCategoryDecision      IncidentEventCategory = "decision"
)

var IncidentEventCategories = []IncidentEventCategory{
	IncidentCategoryStatus,
	IncidentCategoryCommunication,
	IncidentCategoryAction,
	IncidentCategoryObservation,
	IncidentCategoryDecision,
}

func (c IncidentEventCategory) Valid() bool {
	for _, known := range IncidentEventCategories {
		if known == c {
			return true
		}
	}
	return false
}

type IncidentEventSource string

const (
	IncidentSourceSystem IncidentEventSource = "system"
	IncidentSourceUI     IncidentEventSource = "ui"
	IncidentSourceAPI    IncidentEventSource = "api"
	IncidentSourceChat   IncidentEventSource = "chat"
)

func (s IncidentEventSource) Valid() bool {
	switch s {
	case IncidentSourceSystem, IncidentSourceUI, IncidentSourceAPI, IncidentSourceChat:
		return true
	}
	return false
}

func (s IncidentEventSource) Manual() bool {
	return s == IncidentSourceUI || s == IncidentSourceAPI || s == IncidentSourceChat
}

type AttachmentKind string

const (
	AttachmentImage AttachmentKind = "image"
	AttachmentLog   AttachmentKind = "log"
	AttachmentLink  AttachmentKind = "link"
)

func (k AttachmentKind) Valid() bool {
	switch k {
	case AttachmentImage, AttachmentLog, AttachmentLink:
		return true
	}
	return false
}

const (
	TimelineNoteMaxLength    = 4000
	TimelineDefaultPageSize  = 100
	TimelineMaxPageSize      = 200
	TimelineExportMaxEntries = 5000
	TimelineRetroThreshold   = time.Minute
	AttachmentLabelMaxLength = 120
	AttachmentBodyMaxLength  = 20000
	AttachmentURLMaxLength   = 2000
	AttachmentsPerEntryMax   = 10
	AttachmentKeyBytes       = 16
	AttachmentUploadMaxBytes = 10 << 20
)

var (
	ErrTimelineEntryNotFound        = errors.New("timeline entry not found")
	ErrTimelineEntryNotEditable     = errors.New("timeline entry not editable")
	ErrTimelineRetroFuture          = errors.New("timeline entry cannot be in the future")
	ErrTimelineCursorInvalid        = errors.New("timeline cursor invalid")
	ErrAttachmentNotFound           = errors.New("attachment not found")
	ErrAttachmentStorageUnavailable = errors.New("attachment storage unavailable")
	ErrAttachmentTooLarge           = errors.New("attachment too large")
	ErrAttachmentsPerEntryExceeded  = errors.New("attachment limit reached for this entry")
)

type IncidentEventRevision struct {
	ID           string
	EventID      string
	WorkspaceID  string
	At           time.Time
	EditorUserID string
	EditorLabel  string
	Text         string
	Category     IncidentEventCategory
}

type IncidentEventAttachment struct {
	ID          string
	EventID     string
	WorkspaceID string
	Kind        AttachmentKind
	Label       string
	URL         string
	Body        string
	ObjectKey   string
	ContentType string
	SizeBytes   int64
	CreatedAt   time.Time
	CreatedBy   string
}

type NewTimelineEntry struct {
	Category       IncidentEventCategory
	Text           string
	At             time.Time
	Source         IncidentEventSource
	IdempotencyKey string
}

type TimelineEdit struct {
	Text     string
	Category IncidentEventCategory
}

type NewAttachment struct {
	Kind        AttachmentKind
	Label       string
	URL         string
	Body        string
	ContentType string
	SizeBytes   int64
}

type TimelineFilter struct {
	Categories []IncidentEventCategory
	Cursor     string
	Limit      int
}

type TimelinePage struct {
	Entries    []IncidentEvent
	NextCursor string
}

type TimelineCursor struct {
	At time.Time
	ID string
}

func (c TimelineCursor) Zero() bool {
	return c.ID == ""
}

func (c TimelineCursor) Encode() string {
	if c.Zero() {
		return ""
	}
	return c.At.UTC().Format(time.RFC3339Nano) + "|" + c.ID
}

func (c TimelineCursor) Before(other TimelineCursor) bool {
	if c.At.Equal(other.At) {
		return c.ID < other.ID
	}
	return c.At.Before(other.At)
}

func ParseTimelineCursor(raw string) (TimelineCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return TimelineCursor{}, nil
	}
	at, id, ok := strings.Cut(raw, "|")
	if !ok || id == "" {
		return TimelineCursor{}, ErrTimelineCursorInvalid
	}
	parsed, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return TimelineCursor{}, ErrTimelineCursorInvalid
	}
	return TimelineCursor{At: parsed.UTC(), ID: id}, nil
}

func AlertEventKindsForCategories(categories []IncidentEventCategory) []AlertEventKind {
	if len(categories) == 0 {
		return nil
	}
	wanted := map[IncidentEventCategory]bool{}
	for _, c := range categories {
		wanted[c] = true
	}
	out := make([]AlertEventKind, 0, len(AlertEventKinds))
	for _, k := range AlertEventKinds {
		if wanted[CategoryForAlertEvent(k)] {
			out = append(out, k)
		}
	}
	return out
}

type TimelineExport struct {
	Incident   Incident
	Entries    []IncidentEvent
	ExportedAt time.Time
}

func (e TimelineExport) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Incident INC-%d — %s\n", e.Incident.Number, e.Incident.Name)
	fmt.Fprintf(&b, "Severity %s · Status %s\n", e.Incident.SeverityLevel, e.Incident.Status)
	fmt.Fprintf(&b, "Declared %s\n", e.Incident.DeclaredAt.UTC().Format(time.RFC3339))
	if !e.Incident.ResolvedAt.IsZero() {
		fmt.Fprintf(&b, "Resolved %s\n", e.Incident.ResolvedAt.UTC().Format(time.RFC3339))
	}
	fmt.Fprintf(&b, "Exported %s\n", e.ExportedAt.UTC().Format(time.RFC3339))
	b.WriteString("All times UTC\n\n")
	for _, entry := range e.Entries {
		b.WriteString(entry.Line())
		b.WriteString("\n")
		for _, attachment := range entry.Attachments {
			b.WriteString("    " + attachment.Line() + "\n")
			if attachment.Kind == AttachmentLog && attachment.Body != "" {
				for _, line := range strings.Split(attachment.Body, "\n") {
					b.WriteString("        " + line + "\n")
				}
			}
		}
	}
	return b.String()
}

func (e IncidentEvent) Line() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  [%s]  %s", e.At.UTC().Format(time.RFC3339), e.Category, e.Text)
	if e.Actor != "" {
		fmt.Fprintf(&b, " — %s", e.Actor)
	}
	if e.AlertTitle != "" {
		fmt.Fprintf(&b, " — alert %q", e.AlertTitle)
	}
	if e.Result != "" {
		fmt.Fprintf(&b, " (%s)", e.Result)
	}
	if e.Retroactive {
		b.WriteString(" [recorded later]")
	}
	if !e.EditedAt.IsZero() {
		fmt.Fprintf(&b, " [edited %s]", e.EditedAt.UTC().Format(time.RFC3339))
	}
	return b.String()
}

func (a IncidentEventAttachment) Line() string {
	switch a.Kind {
	case AttachmentLink:
		return fmt.Sprintf("link: %s — %s", a.Label, a.URL)
	case AttachmentLog:
		return fmt.Sprintf("log: %s", a.Label)
	default:
		return fmt.Sprintf("image: %s (%d bytes)", a.Label, a.SizeBytes)
	}
}

func AttachmentObjectKey(workspaceID, eventID string) (string, error) {
	suffix, err := GenerateHexToken(AttachmentKeyBytes)
	if err != nil {
		return "", err
	}
	return path.Join("incidents", workspaceID, eventID, suffix), nil
}

func CategoryForKind(kind IncidentEventKind) IncidentEventCategory {
	switch kind {
	case IncidentEventAlertLinked, IncidentEventAlertUnlinked, IncidentEventRelated,
		IncidentEventUnrelated, IncidentEventFollowupAdded, IncidentEventFollowupDone:
		return IncidentCategoryAction
	default:
		return IncidentCategoryStatus
	}
}

func AlertEventOutcome(kind AlertEventKind, result string) string {
	if kind == AlertEventGrouped {
		return ""
	}
	return result
}

func CategoryForAlertEvent(kind AlertEventKind) IncidentEventCategory {
	switch kind {
	case AlertEventAcked, AlertEventResolved:
		return IncidentCategoryStatus
	case AlertEventNotified, AlertEventPush, AlertEventChat:
		return IncidentCategoryCommunication
	case AlertEventEscalation, AlertEventRouted:
		return IncidentCategoryAction
	default:
		return IncidentCategoryObservation
	}
}

func (n NewTimelineEntry) Resolve(now time.Time) (time.Time, bool, error) {
	if n.At.IsZero() {
		return now, false, nil
	}
	at := n.At.UTC()
	if at.After(now) {
		return time.Time{}, false, ErrTimelineRetroFuture
	}
	return at, at.Before(now.Add(-TimelineRetroThreshold)), nil
}

func (n NewTimelineEntry) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Text, validation.By(timelineTextField)),
		validation.Field(&n.Category, validation.By(timelineCategoryField)),
	)
}

func (e TimelineEdit) Validate() error {
	return validation.ValidateStruct(&e,
		validation.Field(&e.Text, validation.By(timelineTextField)),
		validation.Field(&e.Category, validation.By(timelineCategoryField)),
	)
}

func (a NewAttachment) Validate() error {
	if err := validation.ValidateStruct(&a,
		validation.Field(&a.Kind, validation.By(attachmentKindField)),
		validation.Field(&a.Label, validation.By(attachmentLabelField)),
	); err != nil {
		return err
	}
	switch a.Kind {
	case AttachmentLink:
		return validation.ValidateStruct(&a, validation.Field(&a.URL, validation.By(attachmentURLField)))
	case AttachmentLog:
		return validation.ValidateStruct(&a, validation.Field(&a.Body, validation.By(attachmentBodyField)))
	case AttachmentImage:
		return validation.ValidateStruct(&a, validation.Field(&a.ContentType, validation.By(attachmentImageField)))
	}
	return nil
}
