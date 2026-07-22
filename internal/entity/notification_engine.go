package entity

import (
	"strings"
	"time"
)

func NotifyUrgencyFor(severity AlertSeverity) NotifyUrgency {
	if EscalationPriorityLane(severity) == EscalationLaneLow {
		return NotifyUrgencyLow
	}
	return NotifyUrgencyHigh
}

func DefaultNotificationRule(workspaceID, userID string) NotificationRule {
	return NotificationRule{
		WorkspaceID: workspaceID,
		UserID:      userID,
		High:        []NotificationStep{{Channel: ChannelTypeEmail}},
		Low:         []NotificationStep{{Channel: ChannelTypeEmail}},
	}
}

func BuildNotificationPlan(rule NotificationRule, channels []Channel, urgency NotifyUrgency, fallbackEmail string) NotificationPlan {
	steps := rule.High
	if urgency == NotifyUrgencyLow {
		steps = rule.Low
	}
	plan := NotificationPlan{Urgency: urgency, QuietHours: rule.QuietHours}
	for _, step := range steps {
		if ch, ok := resolveChannel(channels, step.Channel); ok {
			plan.Steps = append(plan.Steps, NotificationPlanStep{
				Channel: step.Channel, Delay: step.Delay, ChannelID: ch.ID, Detail: ch.Detail,
			})
			continue
		}
		if step.Channel == ChannelTypeEmail && fallbackEmail != "" {
			plan.Steps = append(plan.Steps, NotificationPlanStep{
				Channel: ChannelTypeEmail, Delay: step.Delay, Detail: fallbackEmail,
			})
		}
	}
	if len(plan.Steps) == 0 && fallbackEmail != "" {
		plan.Steps = []NotificationPlanStep{{Channel: ChannelTypeEmail, Detail: fallbackEmail}}
	}
	return plan
}

func resolveChannel(channels []Channel, kind ChannelType) (Channel, bool) {
	for _, ch := range channels {
		if ch.Type == kind && ch.Verified {
			return ch, true
		}
	}
	return Channel{}, false
}

type NotifyStepKind string

const (
	NotifyStepSend     NotifyStepKind = "send"
	NotifyStepSuppress NotifyStepKind = "suppress"
	NotifyStepExhaust  NotifyStepKind = "exhaust"
)

type NotifyStepTick struct {
	Kind  NotifyStepKind
	Index int
	Step  NotificationPlanStep
}

func (r NotificationRun) NextStep(now time.Time) (NotifyStepTick, bool) {
	if r.State != NotifyRunRunning {
		return NotifyStepTick{}, false
	}
	if now.Before(r.NextAt) {
		return NotifyStepTick{}, false
	}
	if r.StepIndex >= len(r.Plan.Steps) {
		return NotifyStepTick{Kind: NotifyStepExhaust}, true
	}
	step := r.Plan.Steps[r.StepIndex]
	if r.Plan.Suppressed(now) {
		return NotifyStepTick{Kind: NotifyStepSuppress, Index: r.StepIndex, Step: step}, true
	}
	return NotifyStepTick{Kind: NotifyStepSend, Index: r.StepIndex, Step: step}, true
}

func (r NotificationRun) Advanced(now time.Time, outcome NotifyOutcome) NotificationRun {
	out := r
	next := r.StepIndex + 1
	out.StepIndex = next
	if next >= len(r.Plan.Steps) {
		out.State = NotifyRunExhausted
		out.NextAt = time.Time{}
		out.EndedAt = now
		return out
	}
	if outcome == NotifyOutcomeFailed || outcome == NotifyOutcomeSkipped || outcome == NotifyOutcomeThrottled {
		out.NextAt = now
		return out
	}
	scheduled := r.NextAt.Add(r.Plan.Steps[next].Delay)
	if scheduled.Before(now) {
		scheduled = now
	}
	out.NextAt = scheduled
	return out
}

func StartNotificationRun(req NotifyRequest, plan NotificationPlan) NotificationRun {
	return NotificationRun{
		WorkspaceID:     req.WorkspaceID,
		AlertID:         req.AlertID,
		UserID:          req.UserID,
		EscalationID:    req.EscalationID,
		EscalationCycle: req.EscalationCycle,
		Level:           req.Level,
		PolicySlug:      req.PolicySlug,
		Label:           req.Label,
		Urgency:         plan.Urgency,
		State:           NotifyRunRunning,
		Plan:            plan,
		NextAt:          req.At,
		StartedAt:       req.At,
	}
}

func BuildAlertPage(alert Alert, workspaceSlug, policySlug, baseURL string, level int) AlertPage {
	url := strings.TrimRight(baseURL, "/")
	if workspaceSlug != "" {
		url = url + "/" + workspaceSlug + "/alerts/" + alert.ID
	}
	return AlertPage{
		Severity:   alert.Severity,
		Service:    alert.ServiceName,
		Title:      alert.Title,
		StartedAt:  alert.StartedAt,
		PolicySlug: policySlug,
		Level:      level,
		AlertURL:   url,
	}
}
