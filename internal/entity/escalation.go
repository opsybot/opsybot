package entity

import (
	"errors"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type EscalationTargetType string

const (
	EscalationTargetPerson   EscalationTargetType = "person"
	EscalationTargetSchedule EscalationTargetType = "schedule"
	EscalationTargetTeam     EscalationTargetType = "team"
	EscalationTargetWebhook  EscalationTargetType = "webhook"
)

type NotifyMode string

const (
	NotifyModeAll        NotifyMode = "all"
	NotifyModeRoundRobin NotifyMode = "rr"
)

type EscalationBranchKind string

const (
	EscalationBranchPriority EscalationBranchKind = "priority"
	EscalationBranchHours    EscalationBranchKind = "hours"
)

const (
	EscalationLaneHigh     = "high"
	EscalationLaneLow      = "low"
	EscalationLaneBusiness = "business"
	EscalationLaneOff      = "off"
)

type EscalationRunState string

const (
	EscalationRunning   EscalationRunState = "running"
	EscalationAcked     EscalationRunState = "acked"
	EscalationResolved  EscalationRunState = "resolved"
	EscalationExhausted EscalationRunState = "exhausted"
)

const (
	EscalationNameMaxLength   = 60
	EscalationSlugMaxLength   = PolicyRefMaxLength
	EscalationWaitMin         = time.Minute
	EscalationWaitMax         = time.Hour
	EscalationRepeatMax       = 3
	EscalationAckTimeoutMax   = 24 * time.Hour
	EscalationMaxDepth        = 40
	EscalationMaxTargets      = 20
	EscalationSweepBatch      = 100
	EscalationRecentLimit     = 20
	EscalationWebhookNameMax  = 60
	EscalationWebhookURLMax   = 400
	EscalationHoursDayMax     = 6
	EscalationMinutesPerDay   = 24 * 60
	EscalationDefaultStartMin = 9 * 60
	EscalationDefaultEndMin   = 18 * 60
)

var EscalationReservedSlugs = []string{"new"}

type EscalationTarget struct {
	Type EscalationTargetType
	Ref  string
}

type EscalationLevel struct {
	ID      string
	Targets []EscalationTarget
	Mode    NotifyMode
	Wait    time.Duration
}

type EscalationLane struct {
	ID    string
	Key   string
	Nodes []EscalationNode
}

type HoursWindow struct {
	Days        []int
	StartMinute int
	EndMinute   int
	Timezone    string
}

type EscalationBranch struct {
	ID    string
	On    EscalationBranchKind
	Hours HoursWindow
	Lanes []EscalationLane
}

type EscalationNode struct {
	Level  *EscalationLevel
	Branch *EscalationBranch
}

type EscalationPolicy struct {
	ID          string
	WorkspaceID string
	Slug        string
	Name        string
	TeamID      string
	TeamSlug    string
	Repeat      int
	AckTimeout  time.Duration
	Nodes       []EscalationNode
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EscalationPolicySummary struct {
	ID        string
	Slug      string
	Name      string
	TeamSlug  string
	Routed    int
	StepCount int
	HasBranch bool
	Nodes     []EscalationNode
}

type EscalationWebhook struct {
	ID          string
	WorkspaceID string
	Slug        string
	Name        string
	URL         string
	Secret      string
	CreatedAt   time.Time
}

type NewEscalationWebhook struct {
	Slug string
	Name string
	URL  string
}

type EscalationRun struct {
	ID           string
	WorkspaceID  string
	AlertID      string
	PolicyID     string
	PolicySlug   string
	State        EscalationRunState
	Cycle        int
	StepIndex    int
	Snapshot     EscalationSnapshot
	Path         []EscalationLevel
	LaneChoices  map[string]string
	NextAt       time.Time
	AckedAt      time.Time
	AckExpiresAt time.Time
	StartedAt    time.Time
	EndedAt      time.Time
	UpdatedAt    time.Time
}

type EscalationSnapshot struct {
	Repeat     int
	AckTimeout time.Duration
	Nodes      []EscalationNode
}

type EscalationRecent struct {
	AlertID    string
	AlertTitle string
	StartedAt  time.Time
	EndedAt    time.Time
	State      EscalationRunState
	Outcome    string
	ByLabel    string
	StepIndex  int
}

type EscalationDirectory struct {
	Members   []Member
	Schedules []Schedule
	Teams     []Team
	Webhooks  []EscalationWebhook
}

type EscalationPolicyDetail struct {
	Policy    EscalationPolicy
	Routes    []AlertRoute
	Recent    []EscalationRecent
	Routed    int
	IsDefault bool
}

type EscalationPolicyRefs struct {
	Routes     int
	Monitors   int
	Default    bool
	ActiveRuns int
}

var (
	ErrEscalationPolicyNotFound    = errors.New("escalation policy not found")
	ErrEscalationPolicySlugTaken   = errors.New("escalation policy slug taken")
	ErrEscalationPolicyReferenced  = errors.New("escalation policy referenced")
	ErrEscalationPolicyActive      = errors.New("escalation policy has active runs")
	ErrEscalationWebhookNotFound   = errors.New("escalation webhook not found")
	ErrEscalationWebhookSlugTaken  = errors.New("escalation webhook slug taken")
	ErrEscalationWebhookInUse      = errors.New("escalation webhook in use")
	ErrEscalationSecretUnavailable = errors.New("escalation secret storage unavailable")
	ErrEscalationRunNotFound       = errors.New("escalation run not found")
	ErrEscalationRunFinished       = errors.New("escalation run finished")
	ErrEscalationTargetUnknown     = errors.New("escalation target unknown")
	ErrScheduleInPolicy            = errors.New("schedule referenced by an escalation policy")
	ErrTeamInPolicy                = errors.New("team referenced by an escalation policy")
)

func (t EscalationTargetType) Validate() error {
	return escalationTargetTypeField(t)
}

func (m NotifyMode) Validate() error {
	return notifyModeField(m)
}

func (k EscalationBranchKind) Validate() error {
	return escalationBranchKindField(k)
}

func (p EscalationPolicy) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Slug, validation.By(escalationSlugField)),
		validation.Field(&p.Name, validation.By(escalationNameField)),
		validation.Field(&p.Repeat, validation.By(escalationRepeatField)),
		validation.Field(&p.AckTimeout, validation.By(escalationAckTimeoutField)),
		validation.Field(&p.Nodes, validation.By(escalationNodesField)),
	)
}

func (w NewEscalationWebhook) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Slug, validation.By(escalationSlugField)),
		validation.Field(&w.Name, validation.By(escalationWebhookNameField)),
		validation.Field(&w.URL, validation.By(escalationWebhookURLField)),
	)
}

func (p EscalationPolicy) Targets() []EscalationTarget {
	return collectTargets(p.Nodes)
}

func collectTargets(nodes []EscalationNode) []EscalationTarget {
	var out []EscalationTarget
	for _, node := range nodes {
		switch {
		case node.Level != nil:
			out = append(out, node.Level.Targets...)
		case node.Branch != nil:
			for _, lane := range node.Branch.Lanes {
				out = append(out, collectTargets(lane.Nodes)...)
			}
		}
	}
	return out
}

func ValidateEscalationReach(nodes []EscalationNode, known, reachable func(EscalationTarget) bool) error {
	total := 0
	var walk func(nodes []EscalationNode) error
	walk = func(nodes []EscalationNode) error {
		for _, node := range nodes {
			switch {
			case node.Level != nil:
				levelReach := 0
				for _, t := range node.Level.Targets {
					if !known(t) {
						return errEscalationGone
					}
					if reachable(t) {
						levelReach++
						total++
					}
				}
				if len(node.Level.Targets) > 0 && levelReach == 0 {
					return errEscalationDark
				}
			case node.Branch != nil:
				for _, lane := range node.Branch.Lanes {
					if err := walk(lane.Nodes); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	if err := walk(nodes); err != nil {
		return err
	}
	if total == 0 {
		return errEscalationNoReach
	}
	return nil
}

func (r EscalationRun) Active() bool {
	return r.State == EscalationRunning || r.State == EscalationAcked
}

func (r EscalationRun) Outcome() string {
	level := strconv.Itoa(max(r.StepIndex, 1))
	switch r.State {
	case EscalationAcked:
		return "acked at level " + level
	case EscalationResolved:
		return "resolved at level " + level
	case EscalationExhausted:
		return "exhausted: unacked"
	default:
		return "running at level " + level
	}
}

type AlertPage struct {
	Severity   AlertSeverity
	Service    string
	Title      string
	StartedAt  time.Time
	PolicySlug string
	Level      int
	AlertURL   string
	AckURL     string
	ResolveURL string
}

type NotifyResult struct {
	Delivered bool
	Detail    string
	MessageID string
}
