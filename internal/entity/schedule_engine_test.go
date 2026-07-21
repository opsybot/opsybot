package entity

import (
	"testing"
	"time"
)

func mustDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func mustInstant(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func dailyLayer(participants ...string) Layer {
	return Layer{Participants: participants, Rotation: RotationDaily, IntervalDays: 1, HandoverHour: 9, StartsOn: mustDate("2026-07-13")}
}

func assertTiles(t *testing.T, segs []Segment, from, to time.Time) {
	t.Helper()
	if len(segs) == 0 {
		t.Fatal("expected segments, got none")
	}
	if !segs[0].StartsAt.Equal(from) {
		t.Errorf("first segment starts %v, want %v", segs[0].StartsAt.UTC(), from.UTC())
	}
	if !segs[len(segs)-1].EndsAt.Equal(to) {
		t.Errorf("last segment ends %v, want %v", segs[len(segs)-1].EndsAt.UTC(), to.UTC())
	}
	for i := range segs {
		if !segs[i].StartsAt.Before(segs[i].EndsAt) {
			t.Errorf("segment %d is not positive-length: %v..%v", i, segs[i].StartsAt.UTC(), segs[i].EndsAt.UTC())
		}
		if i > 0 && !segs[i].StartsAt.Equal(segs[i-1].EndsAt) {
			t.Errorf("gap or overlap at %d: %v ends, next starts %v", i, segs[i-1].EndsAt.UTC(), segs[i].StartsAt.UTC())
		}
	}
}

func onCall(t *testing.T, s Schedule, instant string) string {
	t.Helper()
	return s.OnCallAt(mustInstant(instant)).UserID
}

func TestRotationDaily(t *testing.T) {
	s := Schedule{Timezone: "UTC", Layers: []Layer{dailyLayer("a", "b", "c")}}
	cases := map[string]string{
		"2026-07-13T09:00:00Z": "a",
		"2026-07-13T23:00:00Z": "a",
		"2026-07-14T09:00:00Z": "b",
		"2026-07-15T12:00:00Z": "c",
		"2026-07-16T12:00:00Z": "a",
	}
	for at, want := range cases {
		if got := onCall(t, s, at); got != want {
			t.Errorf("daily on-call at %s = %q, want %q", at, got, want)
		}
	}
	if before := onCall(t, s, "2026-07-14T08:59:00Z"); before != "a" {
		t.Errorf("just before handover = %q, want a", before)
	}
}

func TestRotationWeeklyAndCustom(t *testing.T) {
	weekly := Schedule{Timezone: "UTC", Layers: []Layer{{Participants: []string{"a", "b"}, Rotation: RotationWeekly, HandoverHour: 9, StartsOn: mustDate("2026-07-13")}}}
	if got := onCall(t, weekly, "2026-07-19T12:00:00Z"); got != "a" {
		t.Errorf("weekly week 0 = %q, want a", got)
	}
	if got := onCall(t, weekly, "2026-07-20T12:00:00Z"); got != "b" {
		t.Errorf("weekly week 1 = %q, want b", got)
	}

	custom := Schedule{Timezone: "UTC", Layers: []Layer{{Participants: []string{"a", "b"}, Rotation: RotationCustom, IntervalDays: 3, HandoverHour: 0, StartsOn: mustDate("2026-07-13")}}}
	if got := onCall(t, custom, "2026-07-15T12:00:00Z"); got != "a" {
		t.Errorf("custom day 2 = %q, want a", got)
	}
	if got := onCall(t, custom, "2026-07-16T12:00:00Z"); got != "b" {
		t.Errorf("custom day 3 = %q, want b", got)
	}
}

func TestCoveragePrecedence(t *testing.T) {
	from := mustInstant("2026-07-13T00:00:00Z")
	wd := int(from.Weekday())
	top := Layer{Participants: []string{"lead"}, Rotation: RotationWeekly, HandoverHour: 0, StartsOn: mustDate("2026-07-13"),
		Restrictions: []Restriction{{Weekday: wd, StartMinute: 9 * 60, EndMinute: 17 * 60}}}
	base := Layer{Participants: []string{"base"}, Rotation: RotationWeekly, HandoverHour: 0, StartsOn: mustDate("2026-07-13")}
	s := Schedule{Timezone: "UTC", Layers: []Layer{top, base}}

	if got := onCall(t, s, "2026-07-13T12:00:00Z"); got != "lead" {
		t.Errorf("inside restriction window = %q, want lead", got)
	}
	if got := onCall(t, s, "2026-07-13T20:00:00Z"); got != "base" {
		t.Errorf("outside restriction window = %q, want base", got)
	}
}

func TestRestrictionGap(t *testing.T) {
	from := mustInstant("2026-07-19T00:00:00Z")
	to := from.Add(24 * time.Hour)
	wd := int(from.Weekday())
	layer := Layer{Participants: []string{"a"}, Rotation: RotationWeekly, HandoverHour: 0, StartsOn: mustDate("2026-07-13"),
		Restrictions: []Restriction{{Weekday: wd, StartMinute: 0, EndMinute: 18 * 60}, {Weekday: wd, StartMinute: 22 * 60, EndMinute: 24 * 60}}}
	s := Schedule{Timezone: "UTC", Layers: []Layer{layer}}

	gaps := s.Gaps(from, to)
	if len(gaps) != 1 {
		t.Fatalf("expected 1 gap, got %d: %+v", len(gaps), gaps)
	}
	if !gaps[0].StartsAt.Equal(from.Add(18*time.Hour)) || !gaps[0].EndsAt.Equal(from.Add(22*time.Hour)) {
		t.Errorf("gap = %v..%v, want 18:00..22:00", gaps[0].StartsAt.UTC(), gaps[0].EndsAt.UTC())
	}
	assertTiles(t, s.Segments(from, to, -1), from, to)
}

func TestOverrides(t *testing.T) {
	s := Schedule{Timezone: "UTC", Layers: []Layer{{Participants: []string{"a", "b"}, Rotation: RotationDaily, HandoverHour: 0, StartsOn: mustDate("2026-07-13")}}}
	s.Overrides = []Override{{UserID: "c", StartsAt: mustInstant("2026-07-13T09:00:00Z"), EndsAt: mustInstant("2026-07-13T12:00:00Z")}}

	if got := onCall(t, s, "2026-07-13T10:00:00Z"); got != "c" {
		t.Errorf("during override = %q, want c", got)
	}
	if got := onCall(t, s, "2026-07-13T13:00:00Z"); got != "a" {
		t.Errorf("after override = %q, want a", got)
	}
	if !s.OverrideConflicts(mustInstant("2026-07-13T11:00:00Z"), mustInstant("2026-07-13T13:00:00Z")) {
		t.Error("overlapping window should conflict")
	}
	if s.OverrideConflicts(mustInstant("2026-07-13T12:00:00Z"), mustInstant("2026-07-13T13:00:00Z")) {
		t.Error("adjacent window should not conflict")
	}
	if !s.OverrideRedundant("a", mustInstant("2026-07-13T13:00:00Z"), mustInstant("2026-07-13T18:00:00Z")) {
		t.Error("assigning the current holder should be redundant")
	}
	if s.OverrideRedundant("b", mustInstant("2026-07-13T13:00:00Z"), mustInstant("2026-07-13T18:00:00Z")) {
		t.Error("assigning a different person should not be redundant")
	}
}

func TestDSTWallClockStable(t *testing.T) {
	s := Schedule{Timezone: "Europe/Berlin", Layers: []Layer{{Participants: []string{"a", "b"}, Rotation: RotationWeekly, HandoverHour: 9, StartsOn: mustDate("2026-03-16")}}}

	handovers := s.HandoverList(mustInstant("2026-03-16T12:00:00Z"), mustInstant("2026-04-07T00:00:00Z"), 5)
	want := map[string]bool{
		"2026-03-23T08:00:00Z": true,
		"2026-03-30T07:00:00Z": true,
		"2026-04-06T07:00:00Z": true,
	}
	for _, h := range handovers {
		delete(want, h.At.UTC().Format(time.RFC3339))
	}
	if len(want) != 0 {
		t.Errorf("missing DST-adjusted handovers at %v; got %+v", want, handovers)
	}
}

func TestSegmentsTileAcrossSpringForward(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("load Berlin: %v", err)
	}
	sunday := int(time.Date(2026, 3, 29, 12, 0, 0, 0, loc).Weekday())
	s := Schedule{Timezone: "Europe/Berlin", Layers: []Layer{{
		Participants: []string{"a", "b"}, Rotation: RotationDaily, HandoverHour: 9, StartsOn: mustDate("2026-03-20"),
		Restrictions: []Restriction{{Weekday: sunday, StartMinute: 8 * 60, EndMinute: 20 * 60}},
	}}}
	from := time.Date(2026, 3, 29, 0, 0, 0, 0, loc).UTC()
	to := time.Date(2026, 3, 30, 0, 0, 0, 0, loc).UTC()
	assertTiles(t, s.Segments(from, to, -1), from, to)
}

func TestSegmentsTileAcrossFallBack(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("load Berlin: %v", err)
	}
	sunday := int(time.Date(2026, 10, 25, 12, 0, 0, 0, loc).Weekday())
	s := Schedule{Timezone: "Europe/Berlin", Layers: []Layer{{
		Participants: []string{"a", "b"}, Rotation: RotationDaily, HandoverHour: 9, StartsOn: mustDate("2026-10-19"),
		Restrictions: []Restriction{{Weekday: sunday, StartMinute: 1 * 60, EndMinute: 5 * 60}},
	}}}
	from := time.Date(2026, 10, 25, 0, 0, 0, 0, loc).UTC()
	to := time.Date(2026, 10, 26, 0, 0, 0, 0, loc).UTC()
	assertTiles(t, s.Segments(from, to, -1), from, to)
}

func TestLeapDayRotation(t *testing.T) {
	daily := Schedule{Timezone: "UTC", Layers: []Layer{{Participants: []string{"a", "b", "c"}, Rotation: RotationDaily, HandoverHour: 0, StartsOn: mustDate("2028-02-27")}}}
	if got := onCall(t, daily, "2028-02-29T12:00:00Z"); got != "c" {
		t.Errorf("leap day on-call = %q, want c", got)
	}
	if got := onCall(t, daily, "2028-03-01T12:00:00Z"); got != "a" {
		t.Errorf("day after leap day = %q, want a", got)
	}

	custom := Schedule{Timezone: "UTC", Layers: []Layer{{Participants: []string{"a", "b"}, Rotation: RotationCustom, IntervalDays: 3, HandoverHour: 0, StartsOn: mustDate("2028-02-26")}}}
	if got := onCall(t, custom, "2028-02-28T12:00:00Z"); got != "a" {
		t.Errorf("custom before leap boundary = %q, want a", got)
	}
	if got := onCall(t, custom, "2028-03-01T12:00:00Z"); got != "b" {
		t.Errorf("custom after leap boundary = %q, want b", got)
	}
}

func TestOnCallPausedAndShifts(t *testing.T) {
	s := Schedule{ID: "s1", Slug: "primary", Timezone: "UTC", Layers: []Layer{dailyLayer("a", "b")}}
	if got := s.OnCallAt(mustInstant("2026-07-13T12:00:00Z")).UserID; got != "a" {
		t.Fatalf("active on-call = %q, want a", got)
	}
	s.Paused = true
	if got := s.OnCallAt(mustInstant("2026-07-13T12:00:00Z")).UserID; got != "" {
		t.Errorf("paused on-call = %q, want empty", got)
	}
	if seg, ok := s.OnCallSegment(mustInstant("2026-07-13T12:00:00Z")); ok {
		t.Errorf("paused OnCallSegment returned %+v", seg)
	}

	s.Paused = false
	shifts := Shifts([]Schedule{s}, "b", mustInstant("2026-07-13T00:00:00Z"), mustInstant("2026-07-16T00:00:00Z"))
	if len(shifts) == 0 {
		t.Fatal("expected shifts for b")
	}
	for _, sh := range shifts {
		if sh.ScheduleSlug != "primary" {
			t.Errorf("shift schedule = %q, want primary", sh.ScheduleSlug)
		}
	}
}

func TestSoloLayerNeverHandsOver(t *testing.T) {
	s := Schedule{Timezone: "UTC", Layers: []Layer{{
		Participants: []string{"a"}, Rotation: RotationCustom, IntervalDays: 2,
		HandoverHour: 9, StartsOn: mustDate("2026-07-13"),
	}}}

	for _, at := range []string{
		"2026-07-13T12:00:00Z",
		"2026-07-15T12:00:00Z",
		"2026-07-17T12:00:00Z",
		"2026-08-02T12:00:00Z",
	} {
		if got := onCall(t, s, at); got != "a" {
			t.Errorf("solo layer on-call at %s = %q, want a", at, got)
		}
	}

	from := mustInstant("2026-07-13T00:00:00Z")
	to := mustInstant("2026-07-27T00:00:00Z")
	if handovers := s.HandoverList(from, to, 5); len(handovers) != 0 {
		t.Errorf("solo layer produced %d handovers, want none: %+v", len(handovers), handovers)
	}
}

func TestCustomIntervalCyclesAllParticipants(t *testing.T) {
	s := Schedule{Timezone: "UTC", Layers: []Layer{{
		Participants: []string{"a", "b", "c"}, Rotation: RotationCustom, IntervalDays: 2,
		HandoverHour: 0, StartsOn: mustDate("2026-07-13"),
	}}}

	cases := map[string]string{
		"2026-07-13T12:00:00Z": "a",
		"2026-07-14T23:59:00Z": "a",
		"2026-07-15T00:00:00Z": "b",
		"2026-07-17T00:00:00Z": "c",
		"2026-07-19T00:00:00Z": "a",
	}
	for at, want := range cases {
		if got := onCall(t, s, at); got != want {
			t.Errorf("custom/2d three-way at %s = %q, want %q", at, got, want)
		}
	}
}

func TestCustomIntervalHandoverHoldsWallClockAcrossDST(t *testing.T) {
	s := Schedule{Timezone: "Europe/Berlin", Layers: []Layer{{
		Participants: []string{"a", "b"}, Rotation: RotationCustom, IntervalDays: 2,
		HandoverHour: 9, StartsOn: mustDate("2026-03-25"),
	}}}

	handovers := s.HandoverList(mustInstant("2026-03-25T00:00:00Z"), mustInstant("2026-04-01T00:00:00Z"), 10)
	want := map[string]bool{
		"2026-03-27T08:00:00Z": true,
		"2026-03-31T07:00:00Z": true,
	}
	for _, h := range handovers {
		delete(want, h.At.UTC().Format(time.RFC3339))
	}
	if len(want) != 0 {
		t.Errorf("missing DST-adjusted custom handovers at %v; got %+v", want, handovers)
	}
}

func TestCustomIntervalBounds(t *testing.T) {
	for _, interval := range []int{LayerMinIntervalDays, LayerMaxIntervalDays} {
		s := Schedule{Timezone: "UTC", Layers: []Layer{{
			Participants: []string{"a", "b"}, Rotation: RotationCustom, IntervalDays: interval,
			HandoverHour: 0, StartsOn: mustDate("2026-07-13"),
		}}}
		start := mustInstant("2026-07-13T12:00:00Z")
		if got := s.OnCallAt(start).UserID; got != "a" {
			t.Errorf("interval %d at start = %q, want a", interval, got)
		}
		next := start.AddDate(0, 0, interval)
		if got := s.OnCallAt(next).UserID; got != "b" {
			t.Errorf("interval %d after one period = %q, want b", interval, got)
		}
	}
}

func TestZeroIntervalFallsBackToMinimum(t *testing.T) {
	s := Schedule{Timezone: "UTC", Layers: []Layer{{
		Participants: []string{"a", "b"}, Rotation: RotationCustom, IntervalDays: 0,
		HandoverHour: 0, StartsOn: mustDate("2026-07-13"),
	}}}
	if got := onCall(t, s, "2026-07-14T12:00:00Z"); got != "b" {
		t.Errorf("zero interval should clamp to %d day: got %q, want b", LayerMinIntervalDays, got)
	}
}
