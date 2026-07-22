package entity

import (
	"crypto/sha256"
	"encoding/hex"
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
