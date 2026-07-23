package entity

import (
	"errors"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type NotifyUrgency string

const (
	NotifyUrgencyHigh NotifyUrgency = "high"
	NotifyUrgencyLow  NotifyUrgency = "low"
)

type NotifyRunState string

const (
	NotifyRunRunning   NotifyRunState = "running"
	NotifyRunStopped   NotifyRunState = "stopped"
	NotifyRunExhausted NotifyRunState = "exhausted"
)

type NotifyStopReason string

const (
	NotifyStopAcked      NotifyStopReason = "acked"
	NotifyStopResolved   NotifyStopReason = "resolved"
	NotifyStopSuperseded NotifyStopReason = "superseded"
)

type NotifyOutcome string

const (
	NotifyOutcomeDelivered  NotifyOutcome = "delivered"
	NotifyOutcomeFailed     NotifyOutcome = "failed"
	NotifyOutcomeSuppressed NotifyOutcome = "suppressed"
	NotifyOutcomeSkipped    NotifyOutcome = "skipped"
	NotifyOutcomeThrottled  NotifyOutcome = "throttled"
)

const (
	NotificationMaxSteps        = 12
	NotificationStepDelayMax    = time.Hour
	NotificationRunSweepBatch   = 200
	NotificationSendConcurrency = 8
	NotificationSymptomMax      = 140
	NotificationRecentLimit     = 50
)

type NotificationStep struct {
	Channel ChannelType
	Delay   time.Duration
}

type QuietHours struct {
	Enabled bool
	Window  HoursWindow
}

type NotificationRule struct {
	WorkspaceID string
	UserID      string
	High        []NotificationStep
	Low         []NotificationStep
	QuietHours  QuietHours
	UpdatedAt   time.Time
}

type NotificationSettings struct {
	Rule     NotificationRule
	Channels []Channel
}

type NotificationPlanStep struct {
	Channel   ChannelType
	Delay     time.Duration
	ChannelID string
	Detail    string
}

type NotificationPlan struct {
	Urgency    NotifyUrgency
	QuietHours QuietHours
	Steps      []NotificationPlanStep
}

type NotificationRun struct {
	ID              string
	WorkspaceID     string
	AlertID         string
	UserID          string
	EscalationID    string
	EscalationCycle int
	Level           int
	PolicySlug      string
	Label           string
	Urgency         NotifyUrgency
	State           NotifyRunState
	StopReason      NotifyStopReason
	StepIndex       int
	Plan            NotificationPlan
	NextAt          time.Time
	StartedAt       time.Time
	EndedAt         time.Time
	UpdatedAt       time.Time
}

type NotificationAttempt struct {
	ID                string
	RunID             string
	WorkspaceID       string
	AlertID           string
	UserID            string
	StepIndex         int
	Channel           ChannelType
	ChannelID         string
	Detail            string
	Outcome           NotifyOutcome
	ProviderMessageID string
	ErrorDetail       string
	At                time.Time
}

type NotifyRequest struct {
	WorkspaceID     string
	AlertID         string
	UserID          string
	Email           string
	Label           string
	PolicySlug      string
	EscalationID    string
	EscalationCycle int
	Level           int
	Severity        AlertSeverity
	At              time.Time
}

type NotifyTarget struct {
	UserID    string
	Name      string
	Channel   ChannelType
	ChannelID string
	Detail    string
	Secret    string
}

type NtfyMessage struct {
	Server   string
	Topic    string
	Token    string
	Title    string
	Body     string
	Priority int
	Click    string
}

var (
	ErrNotificationRuleNotFound = errors.New("notification rule not found")
	ErrNotificationRunNotFound  = errors.New("notification run not found")
)

func (r NotificationRule) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.High, validation.By(notificationStepsField)),
		validation.Field(&r.Low, validation.By(notificationStepsField)),
		validation.Field(&r.QuietHours, validation.By(notificationQuietHoursField)),
	)
}

func ValidateNotificationReach(rule NotificationRule, connected func(ChannelType) bool) error {
	for _, set := range [][]NotificationStep{rule.High, rule.Low} {
		for _, step := range set {
			if !connected(step.Channel) {
				return errNotificationChannel
			}
		}
	}
	return nil
}

func (r NotificationRule) Active() bool {
	return len(r.High) > 0 || len(r.Low) > 0
}

func (p NotificationPlan) Suppressed(at time.Time) bool {
	return p.Urgency == NotifyUrgencyLow && p.QuietHours.Enabled && p.QuietHours.Window.Contains(at)
}

func (r NotificationRun) Stopped(now time.Time, reason NotifyStopReason) NotificationRun {
	out := r
	out.State = NotifyRunStopped
	out.StopReason = reason
	out.NextAt = time.Time{}
	out.EndedAt = now
	return out
}

func (r NotificationRun) Finished(now time.Time) NotificationRun {
	out := r
	out.State = NotifyRunExhausted
	out.NextAt = time.Time{}
	out.EndedAt = now
	return out
}

func NotifyChannelSummary(plan NotificationPlan) string {
	if len(plan.Steps) == 0 {
		return "no channel"
	}
	labels := make([]string, 0, len(plan.Steps))
	for _, step := range plan.Steps {
		labels = append(labels, string(step.Channel))
	}
	return strings.Join(labels, " → ")
}
