package entity

import (
	"testing"
	"time"
)

func quietWindow(tz string) QuietHours {
	return QuietHours{Enabled: true, Window: HoursWindow{
		Days: []int{0, 1, 2, 3, 4, 5, 6}, StartMinute: 22 * 60, EndMinute: 7 * 60, Timezone: tz,
	}}
}

func planStep(kind ChannelType, delay time.Duration) NotificationPlanStep {
	return NotificationPlanStep{Channel: kind, Delay: delay, Detail: "x@y.z"}
}

func runningRun(steps []NotificationPlanStep, urgency NotifyUrgency, quiet QuietHours, next time.Time) NotificationRun {
	return NotificationRun{
		State: NotifyRunRunning, Urgency: urgency, NextAt: next,
		Plan: NotificationPlan{Urgency: urgency, QuietHours: quiet, Steps: steps},
	}
}

func TestNotifyUrgencyMatchesEscalationLane(t *testing.T) {
	cases := map[AlertSeverity]NotifyUrgency{
		SeverityCritical: NotifyUrgencyHigh,
		SeverityHigh:     NotifyUrgencyHigh,
		SeverityWarning:  NotifyUrgencyLow,
	}
	for sev, want := range cases {
		if got := NotifyUrgencyFor(sev); got != want {
			t.Fatalf("severity %s: urgency %s, want %s", sev, got, want)
		}
	}
}

func TestNextStepCoversEveryState(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	steps := []NotificationPlanStep{planStep(ChannelTypeNtfy, 0), planStep(ChannelTypeEmail, 5*time.Minute)}

	notReady := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, now.Add(time.Minute))
	if _, due := notReady.NextStep(now); due {
		t.Fatal("run scheduled in the future should not be due")
	}

	send := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, now)
	tick, due := send.NextStep(now)
	if !due || tick.Kind != NotifyStepSend || tick.Index != 0 {
		t.Fatalf("expected send at index 0, got %+v due=%v", tick, due)
	}

	past := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, now)
	past.StepIndex = 2
	tick, due = past.NextStep(now)
	if !due || tick.Kind != NotifyStepExhaust {
		t.Fatalf("expected exhaust past the end, got %+v", tick)
	}

	stopped := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, now)
	stopped.State = NotifyRunStopped
	if _, due := stopped.NextStep(now); due {
		t.Fatal("stopped run must never be due")
	}
	exhausted := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, now)
	exhausted.State = NotifyRunExhausted
	if _, due := exhausted.NextStep(now); due {
		t.Fatal("exhausted run must never be due")
	}
}

func TestQuietHoursSuppressLowAndNeverHigh(t *testing.T) {
	inWindow := time.Date(2026, 7, 22, 23, 30, 0, 0, time.UTC)
	outWindow := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	steps := []NotificationPlanStep{planStep(ChannelTypeEmail, 0)}

	low := runningRun(steps, NotifyUrgencyLow, quietWindow("UTC"), inWindow)
	tick, _ := low.NextStep(inWindow)
	if tick.Kind != NotifyStepSuppress {
		t.Fatalf("low urgency inside quiet hours should suppress, got %s", tick.Kind)
	}

	high := runningRun(steps, NotifyUrgencyHigh, quietWindow("UTC"), inWindow)
	tick, _ = high.NextStep(inWindow)
	if tick.Kind != NotifyStepSend {
		t.Fatalf("high urgency must break through quiet hours, got %s", tick.Kind)
	}

	lowOut := runningRun(steps, NotifyUrgencyLow, quietWindow("UTC"), outWindow)
	tick, _ = lowOut.NextStep(outWindow)
	if tick.Kind != NotifyStepSend {
		t.Fatalf("low urgency outside quiet hours should send, got %s", tick.Kind)
	}
}

func TestQuietHoursHonoursTimezone(t *testing.T) {
	steps := []NotificationPlanStep{planStep(ChannelTypeEmail, 0)}
	at := time.Date(2026, 7, 22, 21, 30, 0, 0, time.UTC)
	berlin := runningRun(steps, NotifyUrgencyLow, quietWindow("Europe/Berlin"), at)
	tick, _ := berlin.NextStep(at)
	if tick.Kind != NotifyStepSuppress {
		t.Fatalf("21:30 UTC is 23:30 Berlin, inside quiet hours, got %s", tick.Kind)
	}
}

func TestAdvancedSchedulesNextStep(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	steps := []NotificationPlanStep{planStep(ChannelTypeNtfy, 0), planStep(ChannelTypeEmail, 5*time.Minute)}

	run := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, now)
	delivered := run.Advanced(now, NotifyOutcomeDelivered)
	if !delivered.NextAt.Equal(now.Add(5 * time.Minute)) {
		t.Fatalf("delivered step should schedule at now+delay, got %v", delivered.NextAt)
	}

	failed := run.Advanced(now, NotifyOutcomeFailed)
	if !failed.NextAt.Equal(now) {
		t.Fatalf("failed step should pull next to now, got %v", failed.NextAt)
	}

	last := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, now)
	last.StepIndex = 1
	end := last.Advanced(now, NotifyOutcomeDelivered)
	if end.State != NotifyRunExhausted || !end.NextAt.IsZero() {
		t.Fatalf("advancing the last step should exhaust, got state=%s next=%v", end.State, end.NextAt)
	}
}

func TestAdvancedIsDriftFree(t *testing.T) {
	scheduled := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	observed := scheduled.Add(4 * time.Second)
	steps := []NotificationPlanStep{planStep(ChannelTypeNtfy, 0), planStep(ChannelTypeEmail, 5*time.Minute)}
	run := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, scheduled)
	next := run.Advanced(observed, NotifyOutcomeDelivered)
	if !next.NextAt.Equal(scheduled.Add(5 * time.Minute)) {
		t.Fatalf("schedule should derive from NextAt not observed time, got %v", next.NextAt)
	}
}

func TestAdvancedClampsForwardAfterOutage(t *testing.T) {
	scheduled := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	recovered := scheduled.Add(time.Hour)
	steps := []NotificationPlanStep{planStep(ChannelTypeNtfy, 0), planStep(ChannelTypeEmail, time.Minute)}
	run := runningRun(steps, NotifyUrgencyHigh, QuietHours{}, scheduled)
	next := run.Advanced(recovered, NotifyOutcomeDelivered)
	if next.NextAt.Before(recovered) {
		t.Fatalf("a stale schedule must clamp to now, got %v", next.NextAt)
	}
}

func TestEveryNotificationRunTerminates(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	steps := []NotificationPlanStep{
		planStep(ChannelTypeNtfy, 0), planStep(ChannelTypeTelegram, 2*time.Minute), planStep(ChannelTypeEmail, 5*time.Minute),
	}
	outcomes := []NotifyOutcome{NotifyOutcomeDelivered, NotifyOutcomeFailed, NotifyOutcomeSuppressed}
	for _, oc := range outcomes {
		run := runningRun(steps, NotifyUrgencyLow, QuietHours{}, now)
		clock := now
		for i := 0; i < 100; i++ {
			tick, due := run.NextStep(clock)
			if !due {
				clock = run.NextAt
				continue
			}
			if tick.Kind == NotifyStepExhaust {
				run = run.Finished(clock)
				break
			}
			run = run.Advanced(clock, oc)
			clock = clock.Add(time.Second)
		}
		if run.State != NotifyRunExhausted {
			t.Fatalf("outcome %s: run did not terminate, state=%s index=%d", oc, run.State, run.StepIndex)
		}
	}
}

func TestBuildPlanResolvesVerifiedChannels(t *testing.T) {
	channels := []Channel{
		{ID: "c1", Type: ChannelTypeNtfy, Detail: "ntfy.sh/x", Verified: true},
		{ID: "c2", Type: ChannelTypeTelegram, Detail: "@x", Verified: false},
	}
	rule := NotificationRule{
		High: []NotificationStep{
			{Channel: ChannelTypeNtfy},
			{Channel: ChannelTypeTelegram, Delay: 2 * time.Minute},
			{Channel: ChannelTypeEmail, Delay: 5 * time.Minute},
		},
	}
	plan := BuildNotificationPlan(rule, channels, NotifyUrgencyHigh, "on-call@acme.test")
	if len(plan.Steps) != 2 {
		t.Fatalf("expected ntfy + email fallback (telegram unverified dropped), got %d steps", len(plan.Steps))
	}
	if plan.Steps[0].Channel != ChannelTypeNtfy || plan.Steps[0].ChannelID != "c1" {
		t.Fatalf("first step should be the verified ntfy channel, got %+v", plan.Steps[0])
	}
	if plan.Steps[1].Channel != ChannelTypeEmail || plan.Steps[1].Detail != "on-call@acme.test" {
		t.Fatalf("email step should fall back to member email, got %+v", plan.Steps[1])
	}
}

func TestBuildPlanFallsBackToEmailWhenEmpty(t *testing.T) {
	rule := NotificationRule{High: []NotificationStep{{Channel: ChannelTypeSlack}}}
	plan := BuildNotificationPlan(rule, nil, NotifyUrgencyHigh, "on-call@acme.test")
	if len(plan.Steps) != 1 || plan.Steps[0].Channel != ChannelTypeEmail {
		t.Fatalf("unreachable rule should fall back to one email step, got %+v", plan.Steps)
	}
}

func TestBuildPlanEmptyWhenNoEmailFallback(t *testing.T) {
	rule := NotificationRule{High: []NotificationStep{{Channel: ChannelTypeSlack}}}
	plan := BuildNotificationPlan(rule, nil, NotifyUrgencyHigh, "")
	if len(plan.Steps) != 0 {
		t.Fatalf("no reachable channel and no email should yield an empty plan, got %+v", plan.Steps)
	}
}

func TestDefaultNotificationRuleIsOneImmediateEmail(t *testing.T) {
	rule := DefaultNotificationRule("ws-1", "u-1")
	for _, set := range [][]NotificationStep{rule.High, rule.Low} {
		if len(set) != 1 || set[0].Channel != ChannelTypeEmail || set[0].Delay != 0 {
			t.Fatalf("default rule should be one immediate email per lane, got %+v", set)
		}
	}
	if rule.QuietHours.Enabled {
		t.Fatal("default rule should have quiet hours off")
	}
}

func TestNotificationRuleValidation(t *testing.T) {
	valid := NotificationRule{
		High: []NotificationStep{{Channel: ChannelTypeNtfy}, {Channel: ChannelTypeEmail, Delay: 5 * time.Minute}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid rule rejected: %v", err)
	}

	tooMany := make([]NotificationStep, NotificationMaxSteps+1)
	for i := range tooMany {
		tooMany[i] = NotificationStep{Channel: ChannelTypeEmail}
	}
	if err := (NotificationRule{High: tooMany}).Validate(); err == nil {
		t.Fatal("13 steps should be rejected")
	}

	firstDelay := NotificationRule{High: []NotificationStep{{Channel: ChannelTypeEmail, Delay: 5 * time.Minute}}}
	if err := firstDelay.Validate(); err == nil {
		t.Fatal("a delayed first step should be rejected")
	}

	badChannel := NotificationRule{High: []NotificationStep{{Channel: ChannelType("carrier-pigeon")}}}
	if err := badChannel.Validate(); err == nil {
		t.Fatal("an unknown channel should be rejected")
	}

	tooLong := NotificationRule{High: []NotificationStep{{Channel: ChannelTypeEmail}, {Channel: ChannelTypeEmail, Delay: 2 * time.Hour}}}
	if err := tooLong.Validate(); err == nil {
		t.Fatal("a delay over an hour should be rejected")
	}

	badQuiet := NotificationRule{QuietHours: QuietHours{Enabled: true, Window: HoursWindow{Days: nil, StartMinute: 60, EndMinute: 120, Timezone: "UTC"}}}
	if err := badQuiet.Validate(); err == nil {
		t.Fatal("quiet hours with no days should be rejected")
	}

	offQuiet := NotificationRule{QuietHours: QuietHours{Enabled: false, Window: HoursWindow{}}}
	if err := offQuiet.Validate(); err != nil {
		t.Fatalf("disabled quiet hours should pass regardless of window: %v", err)
	}
}

func TestValidateNotificationReach(t *testing.T) {
	rule := NotificationRule{High: []NotificationStep{{Channel: ChannelTypeNtfy}, {Channel: ChannelTypeEmail}}}
	connected := func(t ChannelType) bool { return t == ChannelTypeNtfy || t == ChannelTypeEmail }
	if err := ValidateNotificationReach(rule, connected); err != nil {
		t.Fatalf("reach should pass when all channels connected: %v", err)
	}
	missing := func(t ChannelType) bool { return t == ChannelTypeEmail }
	if err := ValidateNotificationReach(rule, missing); err == nil {
		t.Fatal("reach should fail when a step channel is not connected")
	}
}

func TestChannelTypeEventKind(t *testing.T) {
	cases := map[ChannelType]AlertEventKind{
		ChannelTypeSlack:    AlertEventChat,
		ChannelTypeTeams:    AlertEventChat,
		ChannelTypeDiscord:  AlertEventChat,
		ChannelTypeTelegram: AlertEventChat,
		ChannelTypeNtfy:     AlertEventPush,
		ChannelTypeEmail:    AlertEventNotified,
		ChannelTypeWebhook:  AlertEventNotified,
	}
	for kind, want := range cases {
		if got := kind.EventKind(); got != want {
			t.Fatalf("channel %s: event kind %s, want %s", kind, got, want)
		}
	}
}

func TestNotifyChannelSummary(t *testing.T) {
	plan := NotificationPlan{Steps: []NotificationPlanStep{
		planStep(ChannelTypeNtfy, 0), planStep(ChannelTypeTelegram, time.Minute), planStep(ChannelTypeEmail, 5*time.Minute),
	}}
	if got := NotifyChannelSummary(plan); got != "ntfy → telegram → email" {
		t.Fatalf("unexpected summary %q", got)
	}
	if got := NotifyChannelSummary(NotificationPlan{}); got != "no channel" {
		t.Fatalf("empty plan summary %q", got)
	}
}
