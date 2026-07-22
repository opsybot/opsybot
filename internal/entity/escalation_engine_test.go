package entity

import (
	"strings"
	"testing"
	"time"
)

func level(id string, wait time.Duration, mode NotifyMode, targets ...EscalationTarget) EscalationNode {
	return EscalationNode{Level: &EscalationLevel{ID: id, Targets: targets, Mode: mode, Wait: wait}}
}

func person(ref string) EscalationTarget {
	return EscalationTarget{Type: EscalationTargetPerson, Ref: ref}
}

func hoursBranch(id string, window HoursWindow, business, off []EscalationNode) EscalationNode {
	return EscalationNode{Branch: &EscalationBranch{
		ID: id, On: EscalationBranchHours, Hours: window,
		Lanes: []EscalationLane{
			{ID: id + "-b", Key: EscalationLaneBusiness, Nodes: business},
			{ID: id + "-o", Key: EscalationLaneOff, Nodes: off},
		},
	}}
}

func priorityBranch(id string, high, low []EscalationNode) EscalationNode {
	return EscalationNode{Branch: &EscalationBranch{
		ID: id, On: EscalationBranchPriority,
		Lanes: []EscalationLane{
			{ID: id + "-h", Key: EscalationLaneHigh, Nodes: high},
			{ID: id + "-l", Key: EscalationLaneLow, Nodes: low},
		},
	}}
}

func testPolicy(nodes ...EscalationNode) EscalationPolicy {
	return EscalationPolicy{
		ID: "pol-1", WorkspaceID: "ws-1", Slug: "payments-primary", Name: "payments-primary",
		Repeat: 1, Nodes: nodes,
	}
}

func TestPriorityLaneMapsAlertSeverities(t *testing.T) {
	cases := map[AlertSeverity]string{
		SeverityCritical: EscalationLaneHigh,
		SeverityHigh:     EscalationLaneHigh,
		SeverityWarning:  EscalationLaneLow,
	}
	for severity, want := range cases {
		if got := EscalationPriorityLane(severity); got != want {
			t.Errorf("lane for %s = %q, want %q", severity, got, want)
		}
	}
}

func TestHoursWindowRespectsTimezoneAndDays(t *testing.T) {
	berlin := HoursWindow{Days: []int{1, 2, 3, 4, 5}, StartMinute: 9 * 60, EndMinute: 18 * 60, Timezone: "Europe/Berlin"}

	insideUTC := mustInstant("2026-07-22T08:30:00Z")
	if !berlin.Contains(insideUTC) {
		t.Error("08:30 UTC is 10:30 in Berlin on a Wednesday, inside the window")
	}
	beforeOpen := mustInstant("2026-07-22T06:30:00Z")
	if berlin.Contains(beforeOpen) {
		t.Error("06:30 UTC is 08:30 in Berlin, before the window opens")
	}
	saturday := mustInstant("2026-07-25T10:00:00Z")
	if berlin.Contains(saturday) {
		t.Error("Saturday is not a working day")
	}
}

func TestHoursWindowOvernightRange(t *testing.T) {
	night := HoursWindow{Days: []int{1, 2, 3, 4, 5}, StartMinute: 22 * 60, EndMinute: 6 * 60, Timezone: "UTC"}

	if !night.Contains(mustInstant("2026-07-22T23:00:00Z")) {
		t.Error("23:00 sits inside a 22:00-06:00 overnight window")
	}
	if !night.Contains(mustInstant("2026-07-22T05:00:00Z")) {
		t.Error("05:00 sits inside a 22:00-06:00 overnight window")
	}
	if night.Contains(mustInstant("2026-07-22T12:00:00Z")) {
		t.Error("noon sits outside a 22:00-06:00 overnight window")
	}
}

func TestFlattenPicksLanesAndRecordsChoices(t *testing.T) {
	window := DefaultHoursWindow()
	nodes := []EscalationNode{
		level("l1", 5*time.Minute, NotifyModeAll, person("u1")),
		hoursBranch("b1", window,
			[]EscalationNode{level("l2b", 5*time.Minute, NotifyModeAll, person("u2"))},
			[]EscalationNode{level("l2o", 10*time.Minute, NotifyModeAll, person("u3"))},
		),
	}

	offHours := mustInstant("2026-07-22T02:00:00Z")
	path, choices := FlattenEscalation(nodes, SeverityCritical, offHours)
	if len(path) != 2 || path[1].ID != "l2o" {
		t.Fatalf("path = %v, want l1 then the off-hours lane l2o", ids(path))
	}
	if choices["b1"] != EscalationLaneOff {
		t.Errorf("lane choice = %q, want off", choices["b1"])
	}

	business := mustInstant("2026-07-22T10:00:00Z")
	path, choices = FlattenEscalation(nodes, SeverityCritical, business)
	if len(path) != 2 || path[1].ID != "l2b" {
		t.Fatalf("path = %v, want l1 then the business lane l2b", ids(path))
	}
	if choices["b1"] != EscalationLaneBusiness {
		t.Errorf("lane choice = %q, want business", choices["b1"])
	}
}

func TestFlattenNestedBranches(t *testing.T) {
	nodes := []EscalationNode{
		priorityBranch("b1",
			[]EscalationNode{
				level("h1", 5*time.Minute, NotifyModeAll, person("u1")),
				hoursBranch("b2", DefaultHoursWindow(),
					[]EscalationNode{level("h2b", 5*time.Minute, NotifyModeAll, person("u2"))},
					[]EscalationNode{level("h2o", 5*time.Minute, NotifyModeAll, person("u3"))},
				),
			},
			[]EscalationNode{level("lo1", 15*time.Minute, NotifyModeAll, person("u4"))},
		),
	}

	at := mustInstant("2026-07-22T02:00:00Z")
	path, choices := FlattenEscalation(nodes, SeverityHigh, at)
	if got := strings.Join(ids(path), ","); got != "h1,h2o" {
		t.Fatalf("high-severity off-hours path = %s, want h1,h2o", got)
	}
	if len(choices) != 2 {
		t.Errorf("recorded %d lane choices, want 2", len(choices))
	}

	path, _ = FlattenEscalation(nodes, SeverityWarning, at)
	if got := strings.Join(ids(path), ","); got != "lo1" {
		t.Fatalf("warning path = %s, want lo1", got)
	}
}

func ids(levels []EscalationLevel) []string {
	out := make([]string, len(levels))
	for i, l := range levels {
		out[i] = l.ID
	}
	return out
}

func TestRoundRobinPicksOneRotating(t *testing.T) {
	l := EscalationLevel{Mode: NotifyModeRoundRobin}
	valid := []EscalationTarget{person("u1"), person("u2"), person("u3")}

	for i, want := range []string{"u1", "u2", "u3", "u1"} {
		picked := l.PickTargets(valid, i)
		if len(picked) != 1 || picked[0].Ref != want {
			t.Fatalf("position %d picked %v, want %s", i, picked, want)
		}
	}
}

func TestAllAtOncePicksEveryone(t *testing.T) {
	l := EscalationLevel{Mode: NotifyModeAll}
	valid := []EscalationTarget{person("u1"), person("u2")}
	if picked := l.PickTargets(valid, 7); len(picked) != 2 {
		t.Fatalf("picked %d targets, want all 2 regardless of cursor", len(picked))
	}
	if picked := l.PickTargets(nil, 0); picked != nil {
		t.Fatal("no valid targets must pick no one, not panic")
	}
}

func TestStartRunSnapshotsPolicyAndIsDueImmediately(t *testing.T) {
	now := mustInstant("2026-07-22T10:00:00Z")
	policy := testPolicy(level("l1", 5*time.Minute, NotifyModeAll, person("u1")))
	alert := Alert{ID: "al-1", WorkspaceID: "ws-1", Severity: SeverityCritical}

	run := StartEscalationRun(alert, policy, now)
	if run.State != EscalationRunning || !run.NextAt.Equal(now) {
		t.Fatalf("run = %+v, want running and due immediately", run)
	}
	tick, due := run.NextTick(now)
	if !due || tick.Kind != EscalationTickNotify || tick.Level.ID != "l1" {
		t.Fatalf("tick = %+v due=%v, want notify l1", tick, due)
	}
}

func TestNextTickCoversEveryState(t *testing.T) {
	now := mustInstant("2026-07-22T10:00:00Z")
	base := EscalationRun{
		State:    EscalationRunning,
		Path:     []EscalationLevel{{ID: "l1", Wait: 5 * time.Minute}},
		Snapshot: EscalationSnapshot{Repeat: 1},
		NextAt:   now,
	}

	if tick, due := base.NextTick(now.Add(-time.Second)); due {
		t.Fatalf("run not yet due fired %+v", tick)
	}

	pastPath := base
	pastPath.StepIndex = 1
	if tick, _ := pastPath.NextTick(now); tick.Kind != EscalationTickRepeat {
		t.Fatalf("end of path with repeats left = %v, want repeat", tick.Kind)
	}

	spent := pastPath
	spent.Cycle = 1
	if tick, _ := spent.NextTick(now); tick.Kind != EscalationTickExhaust {
		t.Fatalf("end of path with no repeats left = %v, want exhaust", tick.Kind)
	}

	acked := base
	acked.State = EscalationAcked
	if _, due := acked.NextTick(now); due {
		t.Fatal("acked without expiry must never fire again")
	}

	ackedExpiring := acked
	ackedExpiring.AckExpiresAt = now
	if tick, due := ackedExpiring.NextTick(now); !due || tick.Kind != EscalationTickResume {
		t.Fatalf("expired ack = %+v due=%v, want resume", tick, due)
	}
	if _, due := ackedExpiring.NextTick(now.Add(-time.Second)); due {
		t.Fatal("ack expiry in the future must not resume yet")
	}

	for _, terminal := range []EscalationRunState{EscalationResolved, EscalationExhausted} {
		done := base
		done.State = terminal
		if tick, due := done.NextTick(now); due {
			t.Fatalf("terminal state %s fired %+v", terminal, tick)
		}
	}
}

func TestRepeatedReflattensAndResets(t *testing.T) {
	now := mustInstant("2026-07-22T02:00:00Z")
	nodes := []EscalationNode{
		hoursBranch("b1", DefaultHoursWindow(),
			[]EscalationNode{level("lb", 5*time.Minute, NotifyModeAll, person("u1"))},
			[]EscalationNode{level("lo", 5*time.Minute, NotifyModeAll, person("u2"))},
		),
	}
	run := StartEscalationRun(Alert{Severity: SeverityCritical}, testPolicy(nodes...), now)
	run.StepIndex = 1

	later := mustInstant("2026-07-22T10:30:00Z")
	repeated := run.Repeated(SeverityCritical, later)
	if repeated.Cycle != 1 || repeated.StepIndex != 0 {
		t.Fatalf("repeat = cycle %d step %d, want cycle 1 step 0", repeated.Cycle, repeated.StepIndex)
	}
	if repeated.Path[0].ID != "lb" {
		t.Errorf("repeat at 10:30 took lane %s, want the re-evaluated business lane", repeated.Path[0].ID)
	}
	if run.Path[0].ID != "lo" {
		t.Errorf("original run mutated: %s", run.Path[0].ID)
	}
}

func TestEveryRunTerminates(t *testing.T) {
	now := mustInstant("2026-07-22T10:00:00Z")
	policy := testPolicy(
		level("l1", time.Minute, NotifyModeAll, person("u1")),
		level("l2", time.Minute, NotifyModeRoundRobin, person("u2"), person("u3")),
	)
	policy.Repeat = EscalationRepeatMax

	run := StartEscalationRun(Alert{Severity: SeverityWarning}, policy, now)
	clock := now
	for i := 0; i < 100; i++ {
		tick, due := run.NextTick(clock)
		if !due {
			clock = run.NextAt
			continue
		}
		switch tick.Kind {
		case EscalationTickNotify:
			run.StepIndex++
			run.NextAt = clock.Add(tick.Level.Wait)
		case EscalationTickRepeat:
			run = run.Repeated(SeverityWarning, clock)
		case EscalationTickExhaust:
			run.State = EscalationExhausted
			if run.Cycle != EscalationRepeatMax {
				t.Fatalf("exhausted after %d cycles, want %d", run.Cycle, EscalationRepeatMax)
			}
			return
		}
	}
	t.Fatal("run never reached a terminal state within 100 ticks: an alert would be lost")
}

func TestPolicyValidation(t *testing.T) {
	valid := testPolicy(level("l1", 5*time.Minute, NotifyModeAll, person("u1")))
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid policy rejected: %v", err)
	}

	cases := map[string]EscalationPolicy{
		"no nodes":      testPolicy(),
		"empty level":   testPolicy(level("l1", 5*time.Minute, NotifyModeAll)),
		"bad wait":      testPolicy(level("l1", time.Second, NotifyModeAll, person("u1"))),
		"bad mode":      testPolicy(EscalationNode{Level: &EscalationLevel{ID: "l1", Targets: []EscalationTarget{person("u1")}, Mode: "x", Wait: 5 * time.Minute}}),
		"blank target":  testPolicy(level("l1", 5*time.Minute, NotifyModeAll, EscalationTarget{Type: EscalationTargetPerson, Ref: " "})),
		"bad type":      testPolicy(level("l1", 5*time.Minute, NotifyModeAll, EscalationTarget{Type: "pager", Ref: "x"})),
		"dead-end lane": testPolicy(priorityBranch("b1", []EscalationNode{level("h", 5*time.Minute, NotifyModeAll, person("u1"))}, nil)),
		"branch only no level": {
			ID: "p", Slug: "p-only", Name: "p", Nodes: []EscalationNode{
				priorityBranch("b1",
					[]EscalationNode{priorityBranch("b2", []EscalationNode{level("x", 5*time.Minute, NotifyModeAll, person("u1"))}, []EscalationNode{level("y", 5*time.Minute, NotifyModeAll, person("u2"))})},
					[]EscalationNode{level("z", 5*time.Minute, NotifyModeAll, person("u3"))},
				),
			},
		},
		"bad hours tz": testPolicy(hoursBranch("b1", HoursWindow{Days: []int{1}, StartMinute: 540, EndMinute: 1080, Timezone: "Mars/Olympus"},
			[]EscalationNode{level("a", 5*time.Minute, NotifyModeAll, person("u1"))},
			[]EscalationNode{level("b", 5*time.Minute, NotifyModeAll, person("u2"))},
		)),
		"bad repeat": func() EscalationPolicy {
			p := valid
			p.Repeat = 9
			return p
		}(),
		"reserved slug": func() EscalationPolicy {
			p := valid
			p.Slug = "new"
			return p
		}(),
		"no name": func() EscalationPolicy {
			p := valid
			p.Name = " "
			return p
		}(),
	}
	for name, policy := range cases {
		if name == "branch only no level" {
			if err := policy.Validate(); err != nil {
				t.Errorf("%s: nested branches with levels in every lane should validate, got %v", name, err)
			}
			continue
		}
		if err := policy.Validate(); !IsValidationError(err) {
			t.Errorf("%s: Validate() = %v, want a validation error", name, err)
		}
	}
}

func TestOutcomeStrings(t *testing.T) {
	run := EscalationRun{State: EscalationExhausted, StepIndex: 2}
	if got := run.Outcome(); got != "exhausted: unacked" {
		t.Errorf("outcome = %q", got)
	}
	acked := EscalationRun{State: EscalationAcked, StepIndex: 2}
	if got := acked.Outcome(); got != "acked at level 2" {
		t.Errorf("outcome = %q", got)
	}
}
