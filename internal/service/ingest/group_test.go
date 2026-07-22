package ingest

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/alert_route"
	"github.com/opsybot/opsybot/internal/repository/silence"
)

func withGroupRules(t *testing.T, h *harness, rules []entity.GroupRule) {
	t.Helper()
	ctrl := gomock.NewController(t)
	h.routes = alert_route.NewMockAlertRoute(ctrl)
	h.silences = silence.NewMockSilence(ctrl)
	h.srv.routes = h.routes
	h.srv.silences = h.silences

	h.routes.EXPECT().List(gomock.Any(), "ws-1").Return(nil, nil)
	h.routes.EXPECT().ListGroupRules(gomock.Any(), "ws-1").Return(rules, nil)
	h.routes.EXPECT().Settings(gomock.Any(), "ws-1").
		Return(entity.AlertSettings{DefaultPolicyID: "pol-default", DefaultPolicySlug: "platform-default"}, nil)
	h.silences.EXPECT().ListActive(gomock.Any(), "ws-1", gomock.Any()).Return(nil, nil)
}

func TestWebhookAttachesMatchingAlertsToOneParent(t *testing.T) {
	h := newHarness(t)
	withGroupRules(t, h, []entity.GroupRule{
		{ID: "g1", WorkspaceID: "ws-1", Fields: []string{"service"}, Window: entity.GroupWindowDefault},
	})

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "child-1", Count: 1, ServiceName: "payments-api"}, entity.IngestOutcomeCreated, nil)
	h.alerts.EXPECT().ReplaceLinks(gomock.Any(), "child-1", gomock.Any()).Return(nil)
	h.alerts.EXPECT().UpsertGroupParent(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in entity.AlertUpsert, groupKey string) (entity.Alert, entity.IngestOutcome, error) {
			if groupKey == "" {
				t.Error("group key was empty, so every alert would fall into one parent")
			}
			return entity.Alert{ID: "parent-1", GroupKey: groupKey, Title: in.Title}, entity.IngestOutcomeCreated, nil
		})
	h.alerts.EXPECT().AttachToParent(gomock.Any(), "child-1", "parent-1").Return(nil)
	h.alerts.EXPECT().RollUpParent(gomock.Any(), "parent-1", gomock.Any()).
		Return(entity.Alert{ID: "parent-1", GroupKey: "gk", Count: 3}, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)

	routed := map[string]string{}
	h.alerts.EXPECT().ApplyRouting(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, alertID, policyRef, _, _ string, _ time.Time) error {
			routed[alertID] = policyRef
			return nil
		}).AnyTimes()

	body := `{"title":"payments p99 high","service":"payments-api","severity":"high"}`
	if _, err := h.srv.Webhook(context.Background(), request(body)); err != nil {
		t.Fatalf("webhook: %v", err)
	}

	if routed["child-1"] != "" {
		t.Errorf("child routed to %q, want no policy: only the parent should page", routed["child-1"])
	}
	if routed["parent-1"] != "pol-default" {
		t.Errorf("parent routed to %q, want the default policy id", routed["parent-1"])
	}
}

func TestWebhookLeavesUngroupedAlertsAlone(t *testing.T) {
	h := newHarness(t)
	withGroupRules(t, h, []entity.GroupRule{
		{ID: "g1", WorkspaceID: "ws-1", Fields: []string{"labels.env"}, Window: entity.GroupWindowDefault},
	})

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-1", Count: 1}, entity.IngestOutcomeCreated, nil)
	h.alerts.EXPECT().ReplaceLinks(gomock.Any(), "al-1", gomock.Any()).Return(nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)
	h.allowRouting()

	if _, err := h.srv.Webhook(context.Background(), request(`{"title":"disk full"}`)); err != nil {
		t.Fatalf("webhook: %v", err)
	}
}

func TestFloodRaisesItsOwnAlertAndDropsTheEvent(t *testing.T) {
	h := newHarness(t)
	h.srv.cfg = config.Ingest{MaxBodyBytes: 1 << 20, RatePerMin: 60}

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.limiter.EXPECT().Allow(gomock.Any(), "ingest:src-1", gomock.Any()).
		Return(entity.RateResult{Allowed: false, RetryAfter: time.Minute}, nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in entity.AlertUpsert) (entity.Alert, entity.IngestOutcome, error) {
			if in.DedupKey != entity.FloodDedupPrefix+"src-1" {
				t.Errorf("dedup key = %q, want the reserved flood key so a flood pages once", in.DedupKey)
			}
			return entity.Alert{ID: "flood-1", DedupKey: in.DedupKey}, entity.IngestOutcomeCreated, nil
		})
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "flood-1", gomock.Any()).Return(nil)
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ev entity.IngestEvent) error {
			if ev.Outcome != entity.IngestOutcomeFloodDropped {
				t.Errorf("outcome = %q, want flood_dropped", ev.Outcome)
			}
			return nil
		})

	_, err := h.srv.Webhook(context.Background(), request(`{"title":"disk full"}`))
	if !errors.Is(err, entity.ErrIngestFlooded) {
		t.Fatalf("err = %v, want ErrIngestFlooded", err)
	}
}

func TestUnderBudgetIngestIsNotFlagged(t *testing.T) {
	h := newHarness(t)
	h.srv.cfg = config.Ingest{MaxBodyBytes: 1 << 20, RatePerMin: 60}

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)
	h.limiter.EXPECT().Allow(gomock.Any(), "ingest:src-1", gomock.Any()).
		Return(entity.RateResult{Allowed: true}, nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).
		Return(entity.Alert{ID: "al-1", Count: 1}, entity.IngestOutcomeCreated, nil)
	h.alerts.EXPECT().ReplaceLinks(gomock.Any(), "al-1", gomock.Any()).Return(nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", gomock.Any(), false).Return(nil)
	h.allowRouting()

	if _, err := h.srv.Webhook(context.Background(), request(`{"title":"disk full"}`)); err != nil {
		t.Fatalf("webhook: %v", err)
	}
}
