package ingest

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
)

func heartbeatSource() entity.AlertSource {
	src := genericSource()
	src.Format = entity.SourceFormatHeartbeat
	src.Slug = "nightly-backup"
	src.Name = "nightly-backup"
	return src
}

func testMonitor() entity.AlertMonitor {
	return entity.AlertMonitor{
		ID:          "mon-1",
		WorkspaceID: "ws-1",
		SourceID:    "src-1",
		Slug:        "nightly-backup",
		Name:        "nightly-backup",
		Interval:    time.Hour,
		Grace:       10 * time.Minute,
		Severity:    entity.SeverityHigh,
		PolicyID:    "pol-backups",
		PolicySlug:  "backups-oncall",
		CreatedAt:   time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC),
	}
}

func TestCheckInResolvesTheOpenMissedAlert(t *testing.T) {
	h := newHarness(t)
	src := heartbeatSource()
	monitor := testMonitor()
	now := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(src, nil)
	h.monitors.EXPECT().GetBySourceID(gomock.Any(), "src-1").Return(monitor, nil)
	h.monitors.EXPECT().RecordCheckIn(gomock.Any(), "mon-1", now).Return(nil)
	h.alerts.EXPECT().
		ResolveByDedupKey(gomock.Any(), "ws-1", "src-1", monitor.DedupKey(), now, entity.ResolveModeSource).
		Return(entity.Alert{ID: "al-1"}, entity.IngestOutcomeResolved, nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", now, false).Return(nil)
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)

	got, err := h.srv.CheckIn(context.Background(), entity.CheckInRequest{Token: "tok", ReceivedAt: now})
	if err != nil {
		t.Fatalf("check in: %v", err)
	}
	if got.Outcome != entity.IngestOutcomeResolved || got.AlertID != "al-1" {
		t.Fatalf("got %+v, want the missed alert resolved", got)
	}
}

func TestCheckInWithNothingOpenStillRecordsTheBeat(t *testing.T) {
	h := newHarness(t)
	monitor := testMonitor()
	now := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)

	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(heartbeatSource(), nil)
	h.monitors.EXPECT().GetBySourceID(gomock.Any(), "src-1").Return(monitor, nil)
	h.monitors.EXPECT().RecordCheckIn(gomock.Any(), "mon-1", now).Return(nil)
	h.alerts.EXPECT().ResolveByDedupKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(entity.Alert{}, entity.IngestOutcomeStale, entity.ErrAlertNotFound)
	h.sources.EXPECT().MarkDelivery(gomock.Any(), "src-1", now, false).Return(nil)
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)

	got, err := h.srv.CheckIn(context.Background(), entity.CheckInRequest{Token: "tok", ReceivedAt: now})
	if err != nil {
		t.Fatalf("check in: %v", err)
	}
	if got.Outcome != entity.IngestOutcomeDuplicate {
		t.Fatalf("outcome = %q, want duplicate: a healthy beat has no alert to resolve", got.Outcome)
	}
}

func TestCheckInRejectsAWebhookToken(t *testing.T) {
	h := newHarness(t)
	h.sources.EXPECT().GetByToken(gomock.Any(), "tok").Return(genericSource(), nil)

	_, err := h.srv.CheckIn(context.Background(), entity.CheckInRequest{Token: "tok"})
	if !errors.Is(err, entity.ErrAlertMonitorFormat) {
		t.Fatalf("err = %v, want ErrAlertMonitorFormat: a webhook URL is not a check-in URL", err)
	}
}

func TestSweepRaisesTheMonitorAlertUnderItsOwnPolicy(t *testing.T) {
	h := newHarness(t)
	monitor := testMonitor()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)

	h.monitors.EXPECT().ListDue(gomock.Any(), now, entity.MonitorSweepBatch).
		Return([]entity.AlertMonitor{monitor}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "monitor:mon-1").Return(true, nil)
	h.sources.EXPECT().GetBySlug(gomock.Any(), "ws-1", "nightly-backup").Return(heartbeatSource(), nil)
	h.alerts.EXPECT().UpsertOpen(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in entity.AlertUpsert) (entity.Alert, entity.IngestOutcome, error) {
			if in.DedupKey != monitor.DedupKey() {
				t.Errorf("dedup key = %q, want %q so repeat sweeps do not stack alerts", in.DedupKey, monitor.DedupKey())
			}
			if in.Severity != entity.SeverityHigh {
				t.Errorf("severity = %q, want the monitor severity", in.Severity)
			}
			return entity.Alert{ID: "al-1", Count: 1}, entity.IngestOutcomeCreated, nil
		})
	h.alerts.EXPECT().ReplaceLinks(gomock.Any(), "al-1", gomock.Any()).Return(nil)
	h.alerts.EXPECT().AppendEvent(gomock.Any(), "al-1", gomock.Any()).Return(nil).AnyTimes()
	h.alerts.EXPECT().ApplyRouting(gomock.Any(), "al-1", "pol-backups", "", "", gomock.Any()).Return(nil)
	h.events.EXPECT().Record(gomock.Any(), gomock.Any()).Return(nil)

	fired, err := h.srv.SweepMonitors(context.Background(), now)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if fired != 1 {
		t.Fatalf("fired = %d, want 1", fired)
	}
}

func TestSweepSkipsMonitorsAnotherReplicaHolds(t *testing.T) {
	h := newHarness(t)
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)

	h.monitors.EXPECT().ListDue(gomock.Any(), now, entity.MonitorSweepBatch).
		Return([]entity.AlertMonitor{testMonitor()}, nil)
	h.lock.EXPECT().TryJob(gomock.Any(), "monitor:mon-1").Return(false, nil)

	fired, err := h.srv.SweepMonitors(context.Background(), now)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if fired != 0 {
		t.Fatalf("fired = %d, want 0: another replica already holds this monitor", fired)
	}
}
