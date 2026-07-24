package entity

import (
	"errors"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type IncidentStatus string

const (
	IncidentStatusDeclared      IncidentStatus = "declared"
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusIdentified    IncidentStatus = "identified"
	IncidentStatusMonitoring    IncidentStatus = "monitoring"
	IncidentStatusResolved      IncidentStatus = "resolved"
)

var IncidentStatusOrder = []IncidentStatus{
	IncidentStatusDeclared,
	IncidentStatusInvestigating,
	IncidentStatusIdentified,
	IncidentStatusMonitoring,
	IncidentStatusResolved,
}

func (s IncidentStatus) Valid() bool {
	for _, v := range IncidentStatusOrder {
		if v == s {
			return true
		}
	}
	return false
}

func (s IncidentStatus) Active() bool {
	return s.Valid() && s != IncidentStatusResolved
}

func (s IncidentStatus) index() int {
	for i, v := range IncidentStatusOrder {
		if v == s {
			return i
		}
	}
	return -1
}

// CanTransition reports whether a manual status change from s to target is legal:
// forward exactly one stage, or back one stage (never returning to declared).
// resolved is terminal here — leaving it requires the explicit reopen operation.
func (s IncidentStatus) CanTransition(target IncidentStatus) bool {
	if !s.Valid() || !target.Valid() || s == target || s == IncidentStatusResolved {
		return false
	}
	from, to := s.index(), target.index()
	if to == from+1 {
		return true
	}
	return to == from-1 && target != IncidentStatusDeclared
}

type IncidentRelationKind string

const (
	IncidentRelationRelated   IncidentRelationKind = "related"
	IncidentRelationDuplicate IncidentRelationKind = "duplicate"
	IncidentRelationCausedBy  IncidentRelationKind = "caused_by"
)

func (k IncidentRelationKind) Valid() bool {
	switch k {
	case IncidentRelationRelated, IncidentRelationDuplicate, IncidentRelationCausedBy:
		return true
	}
	return false
}

type CustomFieldKind string

const (
	CustomFieldText        CustomFieldKind = "text"
	CustomFieldSelect      CustomFieldKind = "select"
	CustomFieldMultiSelect CustomFieldKind = "multi_select"
	CustomFieldNumber      CustomFieldKind = "number"
)

func (k CustomFieldKind) Valid() bool {
	switch k {
	case CustomFieldText, CustomFieldSelect, CustomFieldMultiSelect, CustomFieldNumber:
		return true
	}
	return false
}

const (
	IncidentNameMaxLength       = 120
	IncidentSummaryMaxLength    = 2000
	IncidentResolutionMaxLength = 2000
	IncidentListMaxPageSize     = 100
	IncidentTimelineLimit       = 200
	FollowupTitleMaxLength      = 200
	SeverityDefinitionMaxLength = 200
	SeverityLevelMaxLength      = 12
	SeverityLabelMaxLength      = 24
	FieldNameMaxLength          = 60
	FieldOptionsMax             = 40
)

type Incident struct {
	ID                string
	WorkspaceID       string
	Number            int
	Name              string
	Summary           string
	SeverityLevel     string
	Status            IncidentStatus
	LeadUserID        string
	LeadLabel         string
	TeamID            string
	TeamSlug          string
	CustomFields      map[string]string
	ResolutionSummary string
	DeclaredBy        string
	DeclaredAt        time.Time
	ResolvedAt        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Services          []Service
	Alerts            []IncidentAlert
	Related           []IncidentRelation
	Followups         []IncidentFollowup
	Timeline          []IncidentEvent
}

type IncidentAlert struct {
	AlertID  string
	Title    string
	Severity AlertSeverity
	Status   AlertStatus
}

type IncidentRelation struct {
	ID            string
	Kind          IncidentRelationKind
	RelatedID     string
	RelatedNumber int
	RelatedName   string
	RelatedStatus IncidentStatus
}

type IncidentFollowup struct {
	ID          string
	WorkspaceID string
	IncidentID  string
	Title       string
	OwnerUserID string
	OwnerLabel  string
	DueAt       time.Time
	Done        bool
	DoneAt      time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type IncidentEvent struct {
	ID             string
	IncidentID     string
	WorkspaceID    string
	At             time.Time
	Kind           IncidentEventKind
	Category       IncidentEventCategory
	Source         IncidentEventSource
	Text           string
	Actor          string
	ActorUserID    string
	Retroactive    bool
	EditedAt       time.Time
	EditedBy       string
	IdempotencyKey string
	Attachments    []IncidentEventAttachment
	AlertID        string
	AlertTitle     string
	AlertKind      AlertEventKind
	Result         string
}

type IncidentSeverity struct {
	ID          string
	WorkspaceID string
	Level       string
	Label       string
	Definition  string
	Tone        string
	Position    int
}

type IncidentFieldDef struct {
	ID          string
	WorkspaceID string
	Slug        string
	Name        string
	Kind        CustomFieldKind
	Options     []string
	Position    int
}

type IncidentDeclare struct {
	Name          string
	Summary       string
	SeverityLevel string
	TeamSlug      string
	LeadUserID    string
	ServiceIDs    []string
	FromAlertID   string
}

type IncidentUpdate struct {
	Name       string
	Summary    string
	TeamSlug   string
	LeadUserID string
	ServiceIDs []string
}

type IncidentPatch struct {
	Name          *string
	Summary       *string
	SeverityLevel *string
	LeadUserID    *string
	TeamID        *string
}

type NewFollowup struct {
	Title       string
	OwnerUserID string
	DueAt       time.Time
}

type IncidentFilter struct {
	Statuses   []IncidentStatus
	Severities []string
	ServiceIDs []string
	TeamIDs    []string
	ActiveOnly bool
	Since      time.Time
	Query      string
	Cursor     string
	Limit      int
}

type IncidentPage struct {
	Incidents  []Incident
	NextCursor string
}

var (
	ErrIncidentNotFound          = errors.New("incident not found")
	ErrIncidentInvalidTransition = errors.New("incident invalid transition")
	ErrIncidentResolutionMissing = errors.New("incident resolution summary required")
	ErrIncidentSeverityUnknown   = errors.New("incident severity unknown")
	ErrIncidentSelfRelation      = errors.New("incident cannot relate to itself")
	ErrIncidentFieldUnknown      = errors.New("incident custom field unknown")
	ErrIncidentFieldValueInvalid = errors.New("incident custom field value invalid")
	ErrIncidentInvalidCursor     = errors.New("incident invalid cursor")
	ErrIncidentLeadUnknown       = errors.New("incident lead not a member")
	ErrFollowupNotFound          = errors.New("follow-up not found")
	ErrSeverityUnknown           = errors.New("severity level unknown")
	ErrFieldSlugTaken            = errors.New("custom field slug taken")
)

func DefaultSeverities() []IncidentSeverity {
	return []IncidentSeverity{
		{Level: "SEV1", Label: "SEV1", Definition: "Customer-facing outage or data loss. All hands, page immediately.", Tone: "critical", Position: 0},
		{Level: "SEV2", Label: "SEV2", Definition: "Major degradation for many customers. Page the on-call now.", Tone: "high", Position: 1},
		{Level: "SEV3", Label: "SEV3", Definition: "Partial or contained impact. Fix during working hours.", Tone: "warning", Position: 2},
		{Level: "SEV4", Label: "SEV4", Definition: "Minor issue, no customer impact yet. Track it.", Tone: "info", Position: 3},
	}
}

func IncidentSeverityForAlert(sev AlertSeverity) string {
	switch sev {
	case SeverityCritical:
		return "SEV1"
	case SeverityHigh:
		return "SEV2"
	default:
		return "SEV3"
	}
}

func (d IncidentDeclare) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Name, validation.By(incidentNameField)),
		validation.Field(&d.Summary, validation.By(incidentSummaryField)),
	)
}

func (u IncidentUpdate) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.By(incidentNameField)),
		validation.Field(&u.Summary, validation.By(incidentSummaryField)),
	)
}

func (n NewFollowup) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Title, validation.By(followupTitleField)),
	)
}

func (sev IncidentSeverity) Validate() error {
	return validation.ValidateStruct(&sev,
		validation.Field(&sev.Level, validation.By(severityLevelField)),
		validation.Field(&sev.Label, validation.By(severityLabelField)),
		validation.Field(&sev.Definition, validation.By(severityDefinitionField)),
	)
}

func (d IncidentFieldDef) Validate() error {
	if err := validation.ValidateStruct(&d,
		validation.Field(&d.Name, validation.By(fieldNameField)),
		validation.Field(&d.Kind, validation.By(fieldKindField)),
	); err != nil {
		return err
	}
	if d.Kind == CustomFieldSelect || d.Kind == CustomFieldMultiSelect {
		if len(d.Options) == 0 || len(d.Options) > FieldOptionsMax {
			return errFieldOptions
		}
		for _, o := range d.Options {
			if strings.TrimSpace(o) == "" {
				return errFieldOptions
			}
		}
	}
	return nil
}
