package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path"
	"sort"
	"strings"
)

func (a Alert) FieldValue(field string) (string, bool) {
	if key, ok := strings.CutPrefix(field, "labels."); ok {
		v, found := a.Labels[key]
		return v, found
	}
	switch field {
	case "source":
		return a.SourceLabel, true
	case "service":
		return a.ServiceName, true
	case "severity":
		return string(a.Severity), true
	case "title":
		return a.Title, true
	default:
		return "", false
	}
}

func (c RouteCondition) Matches(a Alert) bool {
	actual, present := a.FieldValue(c.Field)
	want := strings.TrimSpace(c.Value)

	switch c.Op {
	case ConditionIs:
		return present && strings.EqualFold(actual, want)
	case ConditionIsNot:
		return !present || !strings.EqualFold(actual, want)
	case ConditionContains:
		return present && strings.Contains(strings.ToLower(actual), strings.ToLower(want))
	case ConditionMatches:
		if !present {
			return false
		}
		ok, err := path.Match(strings.ToLower(want), strings.ToLower(actual))
		return err == nil && ok
	default:
		return false
	}
}

func (r AlertRoute) Matches(a Alert) bool {
	if len(r.Conditions) == 0 {
		return false
	}
	for _, c := range r.Conditions {
		if !c.Matches(a) {
			return false
		}
	}
	return true
}

func RouteFor(routes []AlertRoute, a Alert, defaultPolicyRef string) (AlertRoute, string, bool) {
	ordered := make([]AlertRoute, len(routes))
	copy(ordered, routes)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Position < ordered[j].Position })

	for _, r := range ordered {
		if r.Matches(a) {
			return r, r.PolicyRef, true
		}
	}
	return AlertRoute{}, defaultPolicyRef, false
}

func (g GroupRule) Matches(a Alert) bool {
	for _, field := range g.Fields {
		if v, ok := a.FieldValue(field); !ok || strings.TrimSpace(v) == "" {
			return false
		}
	}
	return len(g.Fields) > 0
}

func PreviewAlert(payload string) (Alert, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return Alert{}, ParseFailure(FailureInvalidJSON, "That sample is not valid JSON.")
	}

	text := func(key string) string {
		s, _ := raw[key].(string)
		return s
	}
	a := Alert{
		Title:       text("title"),
		SourceLabel: text("source"),
		ServiceName: text("service"),
		Severity:    NormalizeSeverity(text("severity"), SeverityWarning),
		Labels:      map[string]string{},
	}
	if labels, ok := raw["labels"].(map[string]any); ok {
		for key, value := range labels {
			if s, ok := value.(string); ok {
				a.Labels[key] = s
			}
		}
	}
	if strings.TrimSpace(a.Title) == "" {
		return Alert{}, ParseFailure(FailureMissingTitle, "That sample needs a title.")
	}
	return a, nil
}

func (g GroupRule) Label(a Alert) string {
	values := make([]string, 0, len(g.Fields))
	for _, field := range g.Fields {
		if v, ok := a.FieldValue(field); ok && strings.TrimSpace(v) != "" {
			values = append(values, v)
		}
	}
	return strings.Join(values, " · ")
}

func (g GroupRule) Describes() string {
	return "Grouped by " + strings.Join(g.Fields, ", ") + ". Matching alerts collapse into this one and it pages once."
}

func GroupParentFor(g GroupRule, groupKey string, child AlertUpsert, a Alert) AlertUpsert {
	labels := make(map[string]string, len(g.Fields))
	for _, field := range g.Fields {
		key, ok := strings.CutPrefix(field, "labels.")
		if !ok {
			continue
		}
		if v, found := a.Labels[key]; found {
			labels[key] = v
		}
	}
	return AlertUpsert{
		WorkspaceID: child.WorkspaceID,
		SourceID:    child.SourceID,
		DedupKey:    GroupDedupPrefix + groupKey,
		Title:       g.Label(a),
		Description: g.Describes(),
		Severity:    child.Severity,
		SourceLabel: child.SourceLabel,
		ServiceName: child.ServiceName,
		Labels:      labels,
		StartedAt:   child.StartedAt,
		LastSeenAt:  child.LastSeenAt,
	}
}

func GroupKeyFor(rules []GroupRule, a Alert) (GroupRule, string, bool) {
	ordered := make([]GroupRule, len(rules))
	copy(ordered, rules)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Position < ordered[j].Position })

	for _, g := range ordered {
		if !g.Matches(a) {
			continue
		}
		parts := make([]string, 0, len(g.Fields)+1)
		parts = append(parts, g.ID)
		for _, field := range g.Fields {
			v, _ := a.FieldValue(field)
			parts = append(parts, field+"="+v)
		}
		sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
		return g, hex.EncodeToString(sum[:]), true
	}
	return GroupRule{}, "", false
}
