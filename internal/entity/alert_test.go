package entity

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeSeverity(t *testing.T) {
	cases := map[string]AlertSeverity{
		"critical": SeverityCritical,
		"FATAL":    SeverityCritical,
		"P1":       SeverityCritical,
		"error":    SeverityHigh,
		"high":     SeverityHigh,
		"warn":     SeverityWarning,
		"info":     SeverityWarning,
		"":         SeverityHigh,
		"nonsense": SeverityHigh,
	}
	for raw, want := range cases {
		if got := NormalizeSeverity(raw, SeverityHigh); got != want {
			t.Errorf("NormalizeSeverity(%q) = %q, want %q", raw, got, want)
		}
	}
	if got := NormalizeSeverity("nonsense", ""); got != SeverityWarning {
		t.Errorf("empty fallback = %q, want warning", got)
	}
}

func TestDeriveDedupKeyPrefersRaw(t *testing.T) {
	got := DeriveDedupKey("src1", "  explicit-key  ", "title", "prom", nil)
	if got != "explicit-key" {
		t.Errorf("raw key = %q, want explicit-key", got)
	}
}

func TestDeriveDedupKeyHashesLongRaw(t *testing.T) {
	long := strings.Repeat("x", DedupKeyMaxLength+1)
	got := DeriveDedupKey("src1", long, "title", "prom", nil)
	if len(got) != 64 {
		t.Errorf("long raw key length = %d, want 64", len(got))
	}
}

func TestDeriveDedupKeyIsStableAndNamespaced(t *testing.T) {
	labels := map[string]string{"env": "prod", "team": "payments"}
	reordered := map[string]string{"team": "payments", "env": "prod"}

	a := DeriveDedupKey("src1", "", "disk full", "prom", labels)
	b := DeriveDedupKey("src1", "", "disk full", "prom", reordered)
	if a != b {
		t.Errorf("label order changed the key: %q vs %q", a, b)
	}

	other := DeriveDedupKey("src2", "", "disk full", "prom", labels)
	if a == other {
		t.Error("different sources produced the same dedup key")
	}
}

func TestTruncateLabelsCaps(t *testing.T) {
	labels := make(map[string]string, AlertLabelsMaxEntries+10)
	for i := range AlertLabelsMaxEntries + 10 {
		labels[string(rune('a'+i%26))+strings.Repeat("k", i)] = "v"
	}
	got := TruncateLabels(labels)
	if len(got) > AlertLabelsMaxEntries {
		t.Errorf("kept %d labels, want at most %d", len(got), AlertLabelsMaxEntries)
	}

	longKey := strings.Repeat("k", AlertLabelKeyMaxLength+50)
	longValue := strings.Repeat("v", AlertLabelValueMaxLength+50)
	one := TruncateLabels(map[string]string{longKey: longValue})
	for k, v := range one {
		if len(k) > AlertLabelKeyMaxLength || len(v) > AlertLabelValueMaxLength {
			t.Errorf("label not truncated: key %d, value %d", len(k), len(v))
		}
	}
}

func TestSourceHealth(t *testing.T) {
	now := mustInstant("2026-07-21T12:00:00Z")

	paused := AlertSource{Paused: true, LastEventAt: now}
	if got := paused.Health(now); got != SourceHealthPaused {
		t.Errorf("paused health = %q, want paused", got)
	}

	failing := AlertSource{FailureCount: 2, LastEventAt: now}
	if got := failing.Health(now); got != SourceHealthFailing {
		t.Errorf("failing health = %q, want failing", got)
	}

	stale := AlertSource{LastEventAt: now.Add(-SourceStaleAfter - 1)}
	if got := stale.Health(now); got != SourceHealthStale {
		t.Errorf("stale health = %q, want stale", got)
	}

	healthy := AlertSource{LastEventAt: now.Add(-1)}
	if got := healthy.Health(now); got != SourceHealthHealthy {
		t.Errorf("healthy health = %q, want healthy", got)
	}
}

func TestSecretsInGraceDropsPreviousAfterWindow(t *testing.T) {
	now := mustInstant("2026-07-21T12:00:00Z")

	fresh := AlertSource{SigningSecret: "new", SigningSecretPrevious: "old", SecretRotatedAt: now.Add(-time.Hour)}
	if got := fresh.SecretsInGrace(now); len(got) != 2 {
		t.Errorf("within grace = %v, want both secrets", got)
	}

	expired := AlertSource{SigningSecret: "new", SigningSecretPrevious: "old", SecretRotatedAt: now.Add(-SourceSecretGrace - time.Hour)}
	got := expired.SecretsInGrace(now)
	if len(got) != 1 || got[0] != "new" {
		t.Errorf("after grace = %v, want only the current secret", got)
	}
}

func TestVerifyBodySignature(t *testing.T) {
	body := []byte(`{"alerts":[]}`)
	sig := SignBody("topsecret", body)

	if !VerifyBodySignature([]string{"topsecret"}, body, sig) {
		t.Error("correct signature rejected")
	}
	if !VerifyBodySignature([]string{"other", "topsecret"}, body, sig) {
		t.Error("signature not matched against the previous secret")
	}
	if VerifyBodySignature([]string{"wrong"}, body, sig) {
		t.Error("wrong secret accepted")
	}
	if VerifyBodySignature([]string{"topsecret"}, []byte(`{"alerts":[1]}`), sig) {
		t.Error("signature accepted for a different body")
	}
	if VerifyBodySignature([]string{"topsecret"}, body, "") {
		t.Error("empty signature accepted")
	}
}

func TestIngestedAlertNormalizeFillsDefaults(t *testing.T) {
	now := mustInstant("2026-07-21T12:00:00Z")
	src := AlertSource{Slug: "prom-prod", DefaultSeverity: SeverityHigh}

	got := IngestedAlert{Title: "  disk full  ", Resolved: true}.Normalize(src, now)

	if got.Title != "disk full" {
		t.Errorf("title = %q, want trimmed", got.Title)
	}
	if got.Severity != SeverityHigh {
		t.Errorf("severity = %q, want the source default", got.Severity)
	}
	if got.SourceLabel != "prom-prod" {
		t.Errorf("source label = %q, want prom-prod", got.SourceLabel)
	}
	if !got.StartedAt.Equal(now) || !got.EndedAt.Equal(now) {
		t.Errorf("timestamps not defaulted: %v / %v", got.StartedAt, got.EndedAt)
	}
	if got.ResolveMode != ResolveModeSource {
		t.Errorf("resolve mode = %q, want source", got.ResolveMode)
	}
}

func TestIngestedAlertValid(t *testing.T) {
	if (IngestedAlert{Title: "  "}).Valid() {
		t.Error("blank title accepted")
	}
	if !(IngestedAlert{Title: "ok"}).Valid() {
		t.Error("valid title rejected")
	}
}
