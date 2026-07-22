package entity

import (
	"testing"
	"time"
)

func sampleAlert() Alert {
	return Alert{
		Title:       "payments-api p99 above 800 ms",
		Severity:    SeverityHigh,
		SourceLabel: "prometheus-prod",
		ServiceName: "payments-api",
		Labels:      map[string]string{"env": "prod", "team": "payments", "region": "eu-west-1"},
	}
}

func TestConditionOperators(t *testing.T) {
	a := sampleAlert()
	cases := []struct {
		name string
		cond RouteCondition
		want bool
	}{
		{"is match", RouteCondition{"service", ConditionIs, "payments-api"}, true},
		{"is case-insensitive", RouteCondition{"service", ConditionIs, "Payments-API"}, true},
		{"is miss", RouteCondition{"service", ConditionIs, "search-api"}, false},
		{"is not match", RouteCondition{"service", ConditionIsNot, "search-api"}, true},
		{"is not miss", RouteCondition{"service", ConditionIsNot, "payments-api"}, false},
		{"contains", RouteCondition{"title", ConditionContains, "p99"}, true},
		{"contains miss", RouteCondition{"title", ConditionContains, "disk"}, false},
		{"matches glob", RouteCondition{"source", ConditionMatches, "prometheus-*"}, true},
		{"matches glob miss", RouteCondition{"source", ConditionMatches, "grafana-*"}, false},
		{"label is", RouteCondition{"labels.env", ConditionIs, "prod"}, true},
		{"label miss", RouteCondition{"labels.env", ConditionIs, "staging"}, false},
		{"absent label is-not passes", RouteCondition{"labels.nope", ConditionIsNot, "x"}, true},
		{"unknown field", RouteCondition{"nonsense", ConditionIs, "x"}, false},
	}
	for _, c := range cases {
		if got := c.cond.Matches(a); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestRouteConditionsAreAnded(t *testing.T) {
	a := sampleAlert()

	both := AlertRoute{PolicyRef: "payments-primary", Conditions: []RouteCondition{
		{"service", ConditionIs, "payments-api"},
		{"labels.env", ConditionIs, "prod"},
	}}
	if !both.Matches(a) {
		t.Error("route with two satisfied conditions did not match")
	}

	one := AlertRoute{PolicyRef: "x", Conditions: []RouteCondition{
		{"service", ConditionIs, "payments-api"},
		{"labels.env", ConditionIs, "staging"},
	}}
	if one.Matches(a) {
		t.Error("route matched with one condition unsatisfied; conditions must be ANDed")
	}

	if (AlertRoute{PolicyRef: "x"}).Matches(a) {
		t.Error("route with no conditions matched")
	}
}

func TestRouteForFirstMatchWinsAndFallsBackToDefault(t *testing.T) {
	a := sampleAlert()
	routes := []AlertRoute{
		{ID: "r2", Position: 2, PolicyRef: "second", Conditions: []RouteCondition{{"labels.env", ConditionIs, "prod"}}},
		{ID: "r1", Position: 1, PolicyRef: "first", Conditions: []RouteCondition{{"service", ConditionIs, "payments-api"}}},
	}

	route, ref, matched := RouteFor(routes, a, "platform-default")
	if !matched || route.ID != "r1" || ref != "first" {
		t.Errorf("got route %q ref %q matched %v, want r1/first by position", route.ID, ref, matched)
	}

	other := sampleAlert()
	other.ServiceName = "search-api"
	other.Labels = map[string]string{}
	_, ref, matched = RouteFor(routes, other, "platform-default")
	if matched || ref != "platform-default" {
		t.Errorf("unmatched alert got ref %q matched %v, want the default route", ref, matched)
	}
}

func TestGroupKeyForIsStableAndDistinct(t *testing.T) {
	rules := []GroupRule{{ID: "g1", Position: 1, Fields: []string{"service", "labels.env"}, Window: GroupWindowDefault}}

	a := sampleAlert()
	_, keyA, ok := GroupKeyFor(rules, a)
	if !ok || keyA == "" {
		t.Fatal("expected a group key")
	}

	same := sampleAlert()
	same.Title = "different title entirely"
	_, keyB, _ := GroupKeyFor(rules, same)
	if keyA != keyB {
		t.Error("group key changed for a field not in the rule")
	}

	other := sampleAlert()
	other.ServiceName = "search-api"
	_, keyC, _ := GroupKeyFor(rules, other)
	if keyA == keyC {
		t.Error("different service produced the same group key")
	}

	missing := sampleAlert()
	missing.Labels = map[string]string{}
	if _, _, matched := GroupKeyFor(rules, missing); matched {
		t.Error("rule matched an alert missing one of its fields")
	}
}

func TestSilenceStateAndMatching(t *testing.T) {
	start := time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	s := Silence{
		StartsAt:   start,
		EndsAt:     end,
		Conditions: []SilenceCondition{{Field: "service", Value: "payments-api"}},
	}

	if got := s.State(start.Add(-time.Minute)); got != SilenceScheduled {
		t.Errorf("before start = %q, want scheduled", got)
	}
	if got := s.State(start.Add(time.Minute)); got != SilenceActive {
		t.Errorf("inside window = %q, want active", got)
	}
	if got := s.State(end); got != SilenceEnded {
		t.Errorf("at end = %q, want ended", got)
	}

	if !s.Matches(sampleAlert()) {
		t.Error("silence did not match the alert it scopes")
	}

	other := sampleAlert()
	other.ServiceName = "search-api"
	if s.Matches(other) {
		t.Error("silence matched an out-of-scope alert")
	}
}

func TestSilenceLabelScope(t *testing.T) {
	s := Silence{Conditions: []SilenceCondition{{Field: "label", Value: "env=prod"}}}
	if !s.Matches(sampleAlert()) {
		t.Error("label scope did not match env=prod")
	}

	bare := Silence{Conditions: []SilenceCondition{{Field: "label", Value: "team"}}}
	if !bare.Matches(sampleAlert()) {
		t.Error("bare label scope did not match on key presence")
	}

	miss := Silence{Conditions: []SilenceCondition{{Field: "label", Value: "env=staging"}}}
	if miss.Matches(sampleAlert()) {
		t.Error("label scope matched the wrong value")
	}
}

func TestSilenceForPicksOnlyActive(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	scoped := []SilenceCondition{{Field: "service", Value: "payments-api"}}

	silences := []Silence{
		{ID: "ended", StartsAt: now.Add(-2 * time.Hour), EndsAt: now.Add(-time.Hour), Conditions: scoped},
		{ID: "scheduled", StartsAt: now.Add(time.Hour), EndsAt: now.Add(2 * time.Hour), Conditions: scoped},
		{ID: "active", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour), Conditions: scoped},
	}

	got, ok := SilenceFor(silences, sampleAlert(), now)
	if !ok || got.ID != "active" {
		t.Errorf("got %q (%v), want the active silence", got.ID, ok)
	}
}
