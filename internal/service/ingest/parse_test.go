package ingest

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return body
}

func at(t *testing.T, iso string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		t.Fatalf("parse %s: %v", iso, err)
	}
	return parsed.UTC()
}

func testSource(format entity.SourceFormat) entity.AlertSource {
	return entity.AlertSource{
		ID:              "src-1",
		Slug:            "prom-prod",
		Format:          format,
		DefaultSeverity: entity.SeverityWarning,
	}
}

func TestParseAlertmanagerDispatchesPerAlert(t *testing.T) {
	now := at(t, "2026-07-21T12:00:00Z")
	got, err := parseFor(entity.SourceFormatAlertmanager, fixture(t, "alertmanager_mixed.json"), testSource(entity.SourceFormatAlertmanager), now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d alerts, want 2 (one firing, one resolved in the same POST)", len(got))
	}

	firing := got[0]
	if firing.Resolved {
		t.Error("first alert resolved, want firing")
	}
	if firing.Severity != entity.SeverityCritical {
		t.Errorf("severity = %q, want critical", firing.Severity)
	}
	if firing.DedupKeyRaw != "aa11bb22" {
		t.Errorf("dedup key = %q, want the fingerprint", firing.DedupKeyRaw)
	}
	if !firing.StartedAt.Equal(at(t, "2026-07-21T10:00:00Z")) {
		t.Errorf("startedAt = %v", firing.StartedAt)
	}
	if !firing.EndedAt.IsZero() {
		t.Errorf("zero endsAt parsed as %v, want zero time", firing.EndedAt)
	}
	if firing.ServiceName != "payments-api" {
		t.Errorf("service = %q, want payments-api", firing.ServiceName)
	}
	if len(firing.Links) < 2 {
		t.Errorf("links = %v, want generator and runbook", firing.Links)
	}

	resolved := got[1]
	if !resolved.Resolved {
		t.Error("second alert firing, want resolved")
	}
	if !resolved.EndedAt.Equal(at(t, "2026-07-21T09:30:00Z")) {
		t.Errorf("resolved endedAt = %v", resolved.EndedAt)
	}
}

func TestParseAlertmanagerRepeatFlag(t *testing.T) {
	now := at(t, "2026-07-21T12:00:00Z")
	got, err := parseFor(entity.SourceFormatAlertmanager, fixture(t, "alertmanager_repeat.json"), testSource(entity.SourceFormatAlertmanager), now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !got[0].Repeat {
		t.Error("repeat interval payload not flagged as a repeat")
	}
}

func TestParseAlertmanagerRejectsOldVersion(t *testing.T) {
	now := at(t, "2026-07-21T12:00:00Z")
	_, err := parseFor(entity.SourceFormatAlertmanager, fixture(t, "alertmanager_v3.json"), testSource(entity.SourceFormatAlertmanager), now)
	pe, ok := entity.ParseFailureOf(err)
	if !ok || pe.Reason != entity.FailureUnsupportedVersion {
		t.Fatalf("v3 payload gave %v, want unsupported_version", err)
	}
}

func TestParseGrafanaStaleReasonResolvesAsTimeout(t *testing.T) {
	now := at(t, "2026-07-21T12:00:00Z")
	got, err := parseFor(entity.SourceFormatGrafana, fixture(t, "grafana_missing_series.json"), testSource(entity.SourceFormatGrafana), now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	alert := got[0]
	if !alert.Resolved {
		t.Fatal("MissingSeries payload not resolved")
	}
	if alert.ResolveMode != entity.ResolveModeTimeout {
		t.Errorf("resolve mode = %q, want timeout so stale series do not look like a real recovery", alert.ResolveMode)
	}
	if !alert.StartedAt.Equal(at(t, "2026-07-21T06:00:00Z")) {
		t.Errorf("startedAt = %v, want the +02:00 offset converted to UTC", alert.StartedAt)
	}
	if _, leaked := alert.Labels["__private__"]; leaked {
		t.Error("__private__ label leaked into user-visible labels")
	}
}

func TestParseKumaDownAndUp(t *testing.T) {
	now := at(t, "2026-07-21T12:00:00Z")
	src := testSource(entity.SourceFormatKuma)

	down, err := parseFor(entity.SourceFormatKuma, fixture(t, "kuma_down.json"), src, now)
	if err != nil {
		t.Fatalf("parse down: %v", err)
	}
	if down[0].Resolved {
		t.Error("status 0 treated as resolved, want firing")
	}
	if !down[0].StartedAt.Equal(at(t, "2026-07-21T14:32:11.128Z")) {
		t.Errorf("kuma naive time = %v, want it read as UTC", down[0].StartedAt)
	}
	if down[0].DedupKeyRaw != "kuma/7" {
		t.Errorf("dedup key = %q, want kuma/7", down[0].DedupKeyRaw)
	}

	up, err := parseFor(entity.SourceFormatKuma, fixture(t, "kuma_up.json"), src, now)
	if err != nil {
		t.Fatalf("parse up: %v", err)
	}
	if !up[0].Resolved {
		t.Error("status 1 not treated as resolved")
	}
	if up[0].DedupKeyRaw != down[0].DedupKeyRaw {
		t.Error("up and down produced different dedup keys, so recovery would not close the alert")
	}
}

func TestParseKumaNullMonitorDegrades(t *testing.T) {
	now := at(t, "2026-07-21T12:00:00Z")
	got, err := parseFor(entity.SourceFormatKuma, fixture(t, "kuma_null_monitor.json"), testSource(entity.SourceFormatKuma), now)
	if err != nil {
		t.Fatalf("null monitor payload errored: %v", err)
	}
	if len(got) != 1 || got[0].Title == "" {
		t.Fatalf("got %+v, want one informational alert", got)
	}
}

func TestParseGenericBatchAndDerivedKey(t *testing.T) {
	now := at(t, "2026-07-21T12:00:00Z")
	src := testSource(entity.SourceFormatGeneric)

	batch, err := parseFor(entity.SourceFormatGeneric, fixture(t, "generic_batch.json"), src, now)
	if err != nil {
		t.Fatalf("parse batch: %v", err)
	}
	if len(batch) != 2 {
		t.Fatalf("got %d alerts, want 2", len(batch))
	}
	if batch[0].Resolved || !batch[1].Resolved {
		t.Error("status field not honoured across the batch")
	}
	if batch[0].DedupKeyRaw != "q-1" || batch[1].DedupKeyRaw != "q-1" {
		t.Error("explicit dedup_key not carried through")
	}

	single, err := parseFor(entity.SourceFormatGeneric, fixture(t, "generic_no_dedup.json"), src, now)
	if err != nil {
		t.Fatalf("parse single: %v", err)
	}
	if single[0].DedupKeyRaw != "" {
		t.Errorf("raw key = %q, want empty so the service derives one", single[0].DedupKeyRaw)
	}
	if !single[0].StartedAt.Equal(now) {
		t.Errorf("startedAt = %v, want the receive time", single[0].StartedAt)
	}
}

func TestParseMalformedPayloadsFailCleanly(t *testing.T) {
	now := at(t, "2026-07-21T12:00:00Z")
	cases := map[string]entity.IngestFailureReason{
		"garbage.txt":    entity.FailureInvalidJSON,
		"truncated.json": entity.FailureInvalidJSON,
		"empty.json":     entity.FailureInvalidJSON,
	}
	for name, wantReason := range cases {
		for _, format := range []entity.SourceFormat{
			entity.SourceFormatAlertmanager,
			entity.SourceFormatGrafana,
			entity.SourceFormatKuma,
			entity.SourceFormatGeneric,
		} {
			got, err := parseFor(format, fixture(t, name), testSource(format), now)
			if err == nil {
				t.Errorf("%s via %s parsed into %+v, want a failure", name, format, got)
				continue
			}
			pe, ok := entity.ParseFailureOf(err)
			if !ok {
				t.Errorf("%s via %s gave %v, want a parse failure", name, format, err)
				continue
			}
			if pe.Reason != wantReason {
				t.Errorf("%s via %s reason = %q, want %q", name, format, pe.Reason, wantReason)
			}
		}
	}
}

func TestParseUnknownFormatIsRejected(t *testing.T) {
	_, err := parseFor(entity.SourceFormat("nope"), []byte(`{}`), testSource("nope"), time.Now())
	pe, ok := entity.ParseFailureOf(err)
	if !ok || pe.Reason != entity.FailureUnsupportedFormat {
		t.Fatalf("unknown format gave %v, want unsupported_format", err)
	}
}
