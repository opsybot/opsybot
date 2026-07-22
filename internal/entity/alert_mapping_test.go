package entity

import (
	"encoding/json"
	"testing"
)

func mustDoc(t *testing.T, raw string) any {
	t.Helper()
	var doc any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return doc
}

func TestLookupPath(t *testing.T) {
	doc := mustDoc(t, `{
		"alerts": [
			{"labels": {"severity": "critical", "env": "prod"}, "count": 3, "ok": true},
			{"labels": {"severity": "warning"}}
		],
		"status": "firing"
	}`)

	cases := map[string]string{
		"status":                    "firing",
		"alerts.0.labels.severity":  "critical",
		"alerts[0].labels.severity": "critical",
		"alerts[1].labels.severity": "warning",
		"alerts[0].count":           "3",
		"alerts[0].ok":              "true",
	}
	for path, want := range cases {
		got, ok := PathString(doc, path)
		if !ok || got != want {
			t.Errorf("PathString(%q) = %q (ok=%v), want %q", path, got, ok, want)
		}
	}

	for _, path := range []string{"missing", "alerts[9].labels", "alerts[0].labels.nope", "status.deeper", ""} {
		if _, ok := PathString(doc, path); ok {
			t.Errorf("PathString(%q) resolved, want miss", path)
		}
	}
}

func TestPathLabels(t *testing.T) {
	doc := mustDoc(t, `{"labels": {"env": "prod", "replicas": 2, "ok": false, "nested": {"a": 1}}}`)

	got := PathLabels(doc, "labels")
	want := map[string]string{"env": "prod", "replicas": "2", "ok": "false"}
	if len(got) != len(want) {
		t.Fatalf("PathLabels = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("label %q = %q, want %q", k, got[k], v)
		}
	}

	if PathLabels(doc, "missing") != nil {
		t.Error("missing path returned labels")
	}
}

func TestMappingPathFor(t *testing.T) {
	mappings := []SourceMapping{
		{Field: MappingFieldTitle, Path: "summary"},
		{Field: MappingFieldSeverity, Path: "labels.severity"},
	}
	if got := MappingPathFor(mappings, MappingFieldSeverity); got != "labels.severity" {
		t.Errorf("severity path = %q, want labels.severity", got)
	}
	if got := MappingPathFor(mappings, MappingFieldService); got != "" {
		t.Errorf("unmapped field = %q, want empty", got)
	}
}

func TestValidateSourceMappings(t *testing.T) {
	if err := ValidateSourceMappings(SourceFormatAlertmanager, nil); err != nil {
		t.Errorf("first-class format required a mapping: %v", err)
	}
	if err := ValidateSourceMappings(SourceFormatGeneric, nil); err == nil {
		t.Error("generic format accepted an empty mapping")
	}

	ok := []SourceMapping{{Field: MappingFieldTitle, Path: "summary"}}
	if err := ValidateSourceMappings(SourceFormatGeneric, ok); err != nil {
		t.Errorf("valid mapping rejected: %v", err)
	}

	noTitle := []SourceMapping{{Field: MappingFieldSeverity, Path: "sev"}}
	if err := ValidateSourceMappings(SourceFormatGeneric, noTitle); err == nil {
		t.Error("mapping without a title accepted")
	}

	dup := []SourceMapping{
		{Field: MappingFieldTitle, Path: "a"},
		{Field: MappingFieldTitle, Path: "b"},
	}
	if err := ValidateSourceMappings(SourceFormatGeneric, dup); err == nil {
		t.Error("duplicate field accepted")
	}

	unknown := []SourceMapping{
		{Field: MappingFieldTitle, Path: "a"},
		{Field: "not_a_field", Path: "b"},
	}
	if err := ValidateSourceMappings(SourceFormatGeneric, unknown); err == nil {
		t.Error("unknown field accepted")
	}
}
