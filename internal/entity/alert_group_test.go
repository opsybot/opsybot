package entity

import (
	"strings"
	"testing"
	"time"
)

func TestGroupParentCarriesOnlyGroupedLabels(t *testing.T) {
	rule := GroupRule{ID: "g1", Fields: []string{"service", "labels.env"}, Window: GroupWindowDefault}
	alert := Alert{
		ServiceName: "payments-api",
		SourceLabel: "prometheus-prod",
		Severity:    SeverityCritical,
		Labels:      map[string]string{"env": "prod", "pod": "payments-7d9f", "team": "payments"},
	}
	child := AlertUpsert{
		WorkspaceID: "ws-1",
		SourceID:    "src-1",
		Severity:    SeverityCritical,
		SourceLabel: "prometheus-prod",
		ServiceName: "payments-api",
		StartedAt:   time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC),
		LastSeenAt:  time.Date(2026, 7, 22, 9, 5, 0, 0, time.UTC),
	}

	parent := GroupParentFor(rule, "abc123", child, alert)

	if parent.DedupKey != GroupDedupPrefix+"abc123" {
		t.Errorf("DedupKey = %q, want the reserved group namespace", parent.DedupKey)
	}
	if parent.Title != "payments-api · prod" {
		t.Errorf("Title = %q, want the grouped field values", parent.Title)
	}
	if _, carried := parent.Labels["pod"]; carried {
		t.Error("parent carried a label it does not group on, so distinct alerts would look identical")
	}
	if parent.Labels["env"] != "prod" {
		t.Errorf("Labels[env] = %q, want prod", parent.Labels["env"])
	}
}

func TestGroupParentIsNotItselfGrouped(t *testing.T) {
	rule := GroupRule{ID: "g1", Fields: []string{"service"}, Window: GroupWindowDefault}
	alert := Alert{ServiceName: "payments-api", Labels: map[string]string{}}
	child := AlertUpsert{WorkspaceID: "ws-1", SourceID: "src-1"}

	parent := GroupParentFor(rule, "abc123", child, alert)
	if !strings.HasPrefix(parent.DedupKey, GroupDedupPrefix) {
		t.Fatalf("DedupKey = %q, want the group prefix that stops the parent regrouping", parent.DedupKey)
	}
}

func TestGroupRuleSkipsAlertsMissingAField(t *testing.T) {
	rules := []GroupRule{{ID: "g1", Fields: []string{"service", "labels.env"}, Window: GroupWindowDefault}}
	withoutEnv := Alert{ServiceName: "payments-api", Labels: map[string]string{"team": "payments"}}

	if _, _, matched := GroupKeyFor(rules, withoutEnv); matched {
		t.Error("grouped an alert missing one of the rule fields, which would merge unrelated alerts")
	}
}

func TestValidateGroupRulesRejectsDuplicateAndUnknownFields(t *testing.T) {
	if err := ValidateGroupRules([]GroupRule{{Fields: []string{"service", "service"}, Window: GroupWindowDefault}}); !IsValidationError(err) {
		t.Errorf("Validate() = %v, want a validation error for a repeated field", err)
	}
	if err := ValidateGroupRules([]GroupRule{{Fields: []string{"nonsense"}, Window: GroupWindowDefault}}); !IsValidationError(err) {
		t.Errorf("Validate() = %v, want a validation error for an unknown field", err)
	}
	if err := ValidateGroupRules([]GroupRule{{Fields: []string{"service"}, Window: time.Second}}); !IsValidationError(err) {
		t.Errorf("Validate() = %v, want a validation error for a one-second window", err)
	}
	if err := ValidateGroupRules([]GroupRule{{Fields: []string{"service", "labels.env"}, Window: GroupWindowDefault}}); err != nil {
		t.Errorf("Validate() = %v, want nil for a valid rule", err)
	}
}
