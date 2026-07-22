package ingest

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/alert"
	"github.com/opsybot/opsybot/internal/repository/alert_monitor"
	"github.com/opsybot/opsybot/internal/repository/alert_route"
	"github.com/opsybot/opsybot/internal/repository/alert_source"
	"github.com/opsybot/opsybot/internal/repository/escalation_policy"
	"github.com/opsybot/opsybot/internal/repository/ingest_event"
	"github.com/opsybot/opsybot/internal/repository/lock"
	"github.com/opsybot/opsybot/internal/repository/ratelimit"
	"github.com/opsybot/opsybot/internal/repository/silence"
	"github.com/opsybot/opsybot/internal/service/escalations"
)

type fakeTx struct{}

func (fakeTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type harness struct {
	srv      *srv
	sources  *alert_source.MockAlertSource
	alerts   *alert.MockAlert
	events   *ingest_event.MockIngestEvent
	routes   *alert_route.MockAlertRoute
	silences *silence.MockSilence
	monitors *alert_monitor.MockAlertMonitor
	limiter  *ratelimit.MockRateLimiter
	lock     *lock.MockLock
	policies *escalation_policy.MockEscalationPolicy
	esc      *escalations.MockEscalations
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	ctrl := gomock.NewController(t)
	h := &harness{
		sources:  alert_source.NewMockAlertSource(ctrl),
		alerts:   alert.NewMockAlert(ctrl),
		events:   ingest_event.NewMockIngestEvent(ctrl),
		routes:   alert_route.NewMockAlertRoute(ctrl),
		silences: silence.NewMockSilence(ctrl),
		monitors: alert_monitor.NewMockAlertMonitor(ctrl),
		limiter:  ratelimit.NewMockRateLimiter(ctrl),
		lock:     lock.NewMockLock(ctrl),
		policies: escalation_policy.NewMockEscalationPolicy(ctrl),
		esc:      escalations.NewMockEscalations(ctrl),
	}
	h.srv = &srv{
		tx:          fakeTx{},
		sources:     h.sources,
		alerts:      h.alerts,
		events:      h.events,
		routes:      h.routes,
		silences:    h.silences,
		monitors:    h.monitors,
		limiter:     h.limiter,
		lock:        h.lock,
		policies:    h.policies,
		escalations: h.esc,
		cfg:         config.Ingest{MaxBodyBytes: 1 << 20},
	}
	h.routes.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	h.routes.EXPECT().ListGroupRules(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	h.routes.EXPECT().Settings(gomock.Any(), gomock.Any()).
		Return(entity.AlertSettings{DefaultPolicyID: "pol-default", DefaultPolicySlug: "platform-default"}, nil).AnyTimes()
	h.silences.EXPECT().ListActive(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	h.policies.EXPECT().List(gomock.Any(), gomock.Any()).
		Return([]entity.EscalationPolicy{{ID: "pol-default", Slug: "platform-default"}}, nil).AnyTimes()
	h.esc.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	return h
}

func (h *harness) allowRouting() {
	h.alerts.EXPECT().ApplyRouting(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
}

func genericSource() entity.AlertSource {
	return entity.AlertSource{
		ID:              "src-1",
		WorkspaceID:     "ws-1",
		Slug:            "generic-prod",
		Format:          entity.SourceFormatGeneric,
		IngestToken:     "tok",
		DefaultSeverity: entity.SeverityWarning,
	}
}

func request(body string) entity.IngestRequest {
	return entity.IngestRequest{
		Token:      "tok",
		Body:       []byte(body),
		ReceivedAt: time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC),
	}
}

func TestWebhookCreatesAlert(t *testing.T) {
	h := newHarness(t)
	src := genericSource()

	h.allowRouting()
	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(src, nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-1", Count: 1}, entity.IngestOutcomeCreated, nil)
	h.alerts.EXPECT().ReplaceLinks(gomock.Any(), "al-1", gomock.Any()).Return(nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil).AnyTimes()
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	got, err := h.srv.Webhook(context.Background(), request(`{"title":"disk full","severity":"critical"}`))
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if len(got) != 1 || got[0].Outcome != entity.IngestOutcomeCreated {
		t.Fatalf("got %+v, want one created result", got)
	}
}

func TestWebhookRepeatIsRecordedAsDuplicate(t *testing.T) {
	h := newHarness(t)
	src := genericSource()
	src.Format = entity.SourceFormatAlertmanager

	h.allowRouting()
	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(src, nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-1", Count: 4}, entity.IngestOutcomeUpdated, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil).AnyTimes()
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ev entity.IngestEvent) error {
			if ev.Outcome != entity.IngestOutcomeDuplicate {
				t.Errorf("recorded outcome %q, want duplicate", ev.Outcome)
			}
			return nil
		})
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	body := string(mustFixture(t, "alertmanager_repeat.json"))
	if _, err := h.srv.Webhook(context.Background(), request(body)); err != nil {
		t.Fatalf("webhook: %v", err)
	}
}

func TestWebhookResolveWithoutOpenAlertInsertsResolved(t *testing.T) {
	h := newHarness(t)

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.alerts.EXPECT().ResolveByDedupKey(gomock.Any(), "ws-1", "src-1", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(entity.Alert{}, entity.IngestOutcomeStale, entity.ErrAlertNotFound)
	h.alerts.EXPECT().FindResolved(gomock.Any(), "ws-1", "src-1", gomock.Any(), gomock.Any()).
		Return(entity.Alert{}, entity.ErrAlertNotFound)
	h.alerts.EXPECT().InsertResolved(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-9"}, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-9", gomock.Any()).Return(nil).AnyTimes()
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	got, err := h.srv.Webhook(context.Background(), request(`{"title":"recovered","status":"resolved","dedup_key":"k9"}`))
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if got[0].Outcome != entity.IngestOutcomeResolved {
		t.Errorf("outcome = %q, want resolved", got[0].Outcome)
	}
}

func TestWebhookRepeatedResolveDoesNotDuplicate(t *testing.T) {
	h := newHarness(t)

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.alerts.EXPECT().ResolveByDedupKey(gomock.Any(), "ws-1", "src-1", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(entity.Alert{}, entity.IngestOutcomeStale, entity.ErrAlertNotFound)
	h.alerts.EXPECT().FindResolved(gomock.Any(), "ws-1", "src-1", gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-existing"}, nil)
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ev entity.IngestEvent) error {
			if ev.Outcome != entity.IngestOutcomeDuplicate {
				t.Errorf("outcome = %q, want duplicate so a re-sent resolve does not create a second row", ev.Outcome)
			}
			return nil
		})
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	got, err := h.srv.Webhook(context.Background(), request(`{"title":"recovered","status":"resolved","dedup_key":"k9","ends_at":"2026-07-21T11:00:00Z"}`))
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if got[0].AlertID != "al-existing" {
		t.Errorf("alert id = %q, want the existing resolved alert", got[0].AlertID)
	}
}

func TestWebhookStaleResolveDoesNotWrite(t *testing.T) {
	h := newHarness(t)

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.alerts.EXPECT().ResolveByDedupKey(gomock.Any(), "ws-1", "src-1", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-2"}, entity.IngestOutcomeStale, nil)
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ev entity.IngestEvent) error {
			if ev.Outcome != entity.IngestOutcomeStale {
				t.Errorf("recorded outcome %q, want stale", ev.Outcome)
			}
			return nil
		})
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	if _, err := h.srv.Webhook(context.Background(), request(`{"title":"old","status":"resolved","dedup_key":"k2"}`)); err != nil {
		t.Fatalf("webhook: %v", err)
	}
}

func TestWebhookMalformedBodyRecordsFailure(t *testing.T) {
	h := newHarness(t)

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.events.EXPECT().RecordFailure(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, f entity.IngestFailure) error {
			if f.Reason != entity.FailureInvalidJSON {
				t.Errorf("reason = %q, want invalid_json", f.Reason)
			}
			if f.Payload != "not json" {
				t.Errorf("payload = %q, want the raw body preserved", f.Payload)
			}
			return nil
		})
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), true).Return(nil)

	_, err := h.srv.Webhook(context.Background(), request("not json"))
	if !errors.Is(err, entity.ErrIngestUnparseable) {
		t.Fatalf("err = %v, want an unparseable ingest error", err)
	}
}

func TestWebhookRejectsPausedSource(t *testing.T) {
	h := newHarness(t)
	src := genericSource()
	src.Paused = true

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(src, nil)

	if _, err := h.srv.Webhook(context.Background(), request(`{"title":"x"}`)); !errors.Is(err, entity.ErrAlertSourcePaused) {
		t.Fatalf("err = %v, want ErrAlertSourcePaused", err)
	}
}

func TestWebhookRejectsBadSignatureWhenPresent(t *testing.T) {
	h := newHarness(t)
	src := genericSource()
	src.SigningSecret = "secret"

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(src, nil)
	h.events.EXPECT().RecordFailure(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), true).Return(nil)

	req := request(`{"title":"x"}`)
	req.Signature = "sha256=deadbeef"

	if _, err := h.srv.Webhook(context.Background(), req); !errors.Is(err, entity.ErrAlertSourceSignature) {
		t.Fatalf("err = %v, want ErrAlertSourceSignature", err)
	}
}

func TestWebhookAcceptsValidSignature(t *testing.T) {
	h := newHarness(t)
	src := genericSource()
	src.SigningSecret = "secret"
	src.RequireSignature = true

	body := `{"title":"signed"}`
	h.allowRouting()
	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(src, nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-3", Count: 1}, entity.IngestOutcomeCreated, nil)
	h.alerts.EXPECT().ReplaceLinks(gomock.Any(), "al-3", gomock.Any()).Return(nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-3", gomock.Any()).Return(nil).AnyTimes()
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	req := request(body)
	req.Signature = entity.SignBody("secret", []byte(body))

	if _, err := h.srv.Webhook(context.Background(), req); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
}

func TestWebhookRejectsOversizeBody(t *testing.T) {
	h := newHarness(t)
	h.srv.cfg.MaxBodyBytes = 8

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.events.EXPECT().RecordFailure(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, f entity.IngestFailure) error {
			if f.Reason != entity.FailureBodyTooLarge {
				t.Errorf("reason = %q, want body_too_large", f.Reason)
			}
			return nil
		})
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), true).Return(nil)

	if _, err := h.srv.Webhook(context.Background(), request(`{"title":"a very long body"}`)); err == nil {
		t.Fatal("oversize body accepted")
	}
}

func mustFixture(t *testing.T, name string) []byte {
	t.Helper()
	return fixture(t, name)
}

func TestWebhookRoutesToMatchingPolicy(t *testing.T) {
	h := newHarness(t)
	ctrl := gomock.NewController(t)
	h.routes = alert_route.NewMockAlertRoute(ctrl)
	h.silences = silence.NewMockSilence(ctrl)
	h.srv.routes = h.routes
	h.srv.silences = h.silences

	h.routes.EXPECT().List(gomock.Any(), "ws-1").Return([]entity.AlertRoute{
		{ID: "r1", Position: 0, PolicyID: "pol-payments", PolicySlug: "payments-primary", Conditions: []entity.RouteCondition{
			{Field: "service", Op: entity.ConditionIs, Value: "payments-api"},
		}},
	}, nil)
	h.routes.EXPECT().ListGroupRules(gomock.Any(), "ws-1").Return(nil, nil)
	h.routes.EXPECT().Settings(gomock.Any(), "ws-1").
		Return(entity.AlertSettings{DefaultPolicyID: "pol-default", DefaultPolicySlug: "platform-default"}, nil)
	h.silences.EXPECT().ListActive(gomock.Any(), "ws-1", gomock.Any()).Return(nil, nil)

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-1", ServiceName: "payments-api", Count: 1}, entity.IngestOutcomeCreated, nil)
	h.alerts.EXPECT().ReplaceLinks(gomock.Any(), "al-1", gomock.Any()).Return(nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil).AnyTimes()
	h.alerts.EXPECT().ApplyRouting(gomock.Any(), "al-1", "pol-payments", "", "", gomock.Any()).Return(nil)
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	if _, err := h.srv.Webhook(context.Background(), request(`{"title":"p99 high","service":"payments-api"}`)); err != nil {
		t.Fatalf("webhook: %v", err)
	}
}

func TestWebhookRecordsSuppressionButStillCreatesAlert(t *testing.T) {
	h := newHarness(t)
	ctrl := gomock.NewController(t)
	h.routes = alert_route.NewMockAlertRoute(ctrl)
	h.silences = silence.NewMockSilence(ctrl)
	h.srv.routes = h.routes
	h.srv.silences = h.silences

	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	h.routes.EXPECT().List(gomock.Any(), "ws-1").Return(nil, nil)
	h.routes.EXPECT().ListGroupRules(gomock.Any(), "ws-1").Return(nil, nil)
	h.routes.EXPECT().Settings(gomock.Any(), "ws-1").
		Return(entity.AlertSettings{DefaultPolicyID: "pol-default", DefaultPolicySlug: "platform-default"}, nil)
	h.silences.EXPECT().ListActive(gomock.Any(), "ws-1", gomock.Any()).Return([]entity.Silence{
		{
			ID:         "sil-1",
			StartsAt:   now.Add(-time.Hour),
			EndsAt:     now.Add(time.Hour),
			Conditions: []entity.SilenceCondition{{Field: "service", Value: "payments-api"}},
		},
	}, nil)

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-2", ServiceName: "payments-api", Count: 1}, entity.IngestOutcomeCreated, nil)
	h.alerts.EXPECT().ReplaceLinks(gomock.Any(), "al-2", gomock.Any()).Return(nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-2", gomock.Any()).Return(nil).AnyTimes()
	h.alerts.EXPECT().ApplyRouting(gomock.Any(), "al-2", "pol-default", "", "sil-1", gomock.Any()).Return(nil)
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	got, err := h.srv.Webhook(context.Background(), request(`{"title":"p99 high","service":"payments-api"}`))
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if got[0].Outcome != entity.IngestOutcomeCreated {
		t.Errorf("outcome = %q, want created: a silenced alert is still recorded", got[0].Outcome)
	}
}
