package incidents

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
)

func at(offset int) time.Time {
	return time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC).Add(time.Duration(offset) * time.Minute)
}

func TestEditRejectsAutomaticEntry(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventStatusChanged, Text: "Status moved to identified",
	}, nil)

	_, err := h.srv.EditTimelineEntry(adminCtx(), "acme", "inc-1", "ev-1", entity.TimelineEdit{
		Text: "Rewritten", Category: entity.IncidentCategoryStatus,
	})
	if !errors.Is(err, entity.ErrTimelineEntryNotEditable) {
		t.Fatalf("err = %v, want ErrTimelineEntryNotEditable", err)
	}
}

func TestEditRecordsPreviousTextAsRevision(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
		Category: entity.IncidentCategoryObservation, Text: "Cache hit rate dropped",
	}, nil)

	var revision entity.IncidentEventRevision
	h.incidents.EXPECT().AppendRevision(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in entity.IncidentEventRevision) error {
			revision = in
			return nil
		})

	var edit entity.TimelineEdit
	h.incidents.EXPECT().UpdateEvent(gomock.Any(), "ws-1", "ev-1", gomock.Any(), gomock.Any(), "u1").DoAndReturn(
		func(_ context.Context, _, _ string, in entity.TimelineEdit, _ time.Time, _ string) error {
			edit = in
			return nil
		})
	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
		Category: entity.IncidentCategoryDecision, Text: "Cache hit rate dropped to 12 percent",
	}, nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	out, err := h.srv.EditTimelineEntry(adminCtx(), "acme", "inc-1", "ev-1", entity.TimelineEdit{
		Text: "Cache hit rate dropped to 12 percent", Category: entity.IncidentCategoryDecision,
	})
	if err != nil {
		t.Fatalf("edit entry: %v", err)
	}
	if revision.Text != "Cache hit rate dropped" {
		t.Errorf("revision keeps %q, want the previous text", revision.Text)
	}
	if revision.Category != entity.IncidentCategoryObservation {
		t.Errorf("revision category = %q, want the previous category", revision.Category)
	}
	if revision.EditorUserID != "u1" || revision.EditorLabel != "Priya" {
		t.Errorf("revision editor = %q/%q, want u1/Priya", revision.EditorUserID, revision.EditorLabel)
	}
	if edit.Text != "Cache hit rate dropped to 12 percent" {
		t.Errorf("stored text = %q", edit.Text)
	}
	if out.Text != "Cache hit rate dropped to 12 percent" {
		t.Errorf("returned text = %q", out.Text)
	}
}

func TestAddEntryRejectsFutureTimestamp(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	_, err := h.srv.AddTimelineEntry(adminCtx(), "acme", "inc-1", entity.NewTimelineEntry{
		Text: "Rolled back the deploy", At: time.Now().UTC().Add(time.Hour),
	})
	if !errors.Is(err, entity.ErrTimelineRetroFuture) {
		t.Fatalf("err = %v, want ErrTimelineRetroFuture", err)
	}
}

func TestAddEntryMarksBackdatedEntryRetroactive(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").Return(entity.Incident{ID: "inc-1", WorkspaceID: "ws-1"}, nil)

	var stored entity.IncidentEvent
	h.incidents.EXPECT().AppendEvent(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in entity.IncidentEvent) (entity.IncidentEvent, error) {
			stored = in
			stored.ID = "ev-9"
			return stored, nil
		})
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	backdated := time.Now().UTC().Add(-2 * time.Hour)
	out, err := h.srv.AddTimelineEntry(adminCtx(), "acme", "inc-1", entity.NewTimelineEntry{
		Text: "Noticed the queue backing up", At: backdated, IdempotencyKey: " chat-42 ",
	})
	if err != nil {
		t.Fatalf("add entry: %v", err)
	}
	if !stored.Retroactive {
		t.Error("backdated entry not marked retroactive")
	}
	if !stored.At.Equal(backdated) {
		t.Errorf("stored at = %s, want %s", stored.At, backdated)
	}
	if stored.Kind != entity.IncidentEventNote || stored.Category != entity.IncidentCategoryObservation {
		t.Errorf("kind/category = %q/%q", stored.Kind, stored.Category)
	}
	if stored.IdempotencyKey != "chat-42" {
		t.Errorf("idempotency key = %q, want it trimmed", stored.IdempotencyKey)
	}
	if out.ID != "ev-9" {
		t.Errorf("returned id = %q", out.ID)
	}
}

func TestTimelineMergesLinkedAlertEvents(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	inc := entity.Incident{
		ID: "inc-1", WorkspaceID: "ws-1",
		Alerts: []entity.IncidentAlert{{AlertID: "al-1", Title: "Checkout latency spike"}},
	}
	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").Return(inc, nil)
	h.incidents.EXPECT().ListEvents(gomock.Any(), "ws-1", "inc-1", gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]entity.IncidentEvent{
			{ID: "a", At: at(0), Kind: entity.IncidentEventDeclared, Text: "Declared at SEV1"},
			{ID: "c", At: at(4), Kind: entity.IncidentEventStatusChanged, Text: "Status moved to identified"},
		}, nil)
	h.alerts.EXPECT().ListEventsForAlerts(gomock.Any(), []string{"al-1"}, gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]entity.AlertEvent{
			{ID: "b", AlertID: "al-1", At: at(2), Kind: entity.AlertEventNotified, Text: "Paged the on-call", Result: "delivered"},
		}, nil)

	page, err := h.srv.ListTimeline(adminCtx(), "acme", "inc-1", entity.TimelineFilter{})
	if err != nil {
		t.Fatalf("list timeline: %v", err)
	}
	got := make([]string, 0, len(page.Entries))
	for _, e := range page.Entries {
		got = append(got, e.ID)
	}
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("merged order = %v, want [a b c]", got)
	}
	merged := page.Entries[1]
	if merged.AlertTitle != "Checkout latency spike" || merged.AlertID != "al-1" {
		t.Errorf("alert attribution missing: %+v", merged)
	}
	if merged.Category != entity.IncidentCategoryCommunication {
		t.Errorf("category = %q, want communication", merged.Category)
	}
	if merged.Result != "delivered" {
		t.Errorf("result = %q, want delivered", merged.Result)
	}
}

func TestTimelineOrdersEqualTimestampsById(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	inc := entity.Incident{ID: "inc-1", WorkspaceID: "ws-1", Alerts: []entity.IncidentAlert{{AlertID: "al-1"}}}
	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").Return(inc, nil)
	h.incidents.EXPECT().ListEvents(gomock.Any(), "ws-1", "inc-1", gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]entity.IncidentEvent{{ID: "0192-b", At: at(3), Kind: entity.IncidentEventNote, Text: "note"}}, nil)
	h.alerts.EXPECT().ListEventsForAlerts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]entity.AlertEvent{{ID: "0192-a", AlertID: "al-1", At: at(3), Kind: entity.AlertEventAcked, Text: "acked"}}, nil)

	page, err := h.srv.ListTimeline(adminCtx(), "acme", "inc-1", entity.TimelineFilter{})
	if err != nil {
		t.Fatalf("list timeline: %v", err)
	}
	if page.Entries[0].ID != "0192-a" || page.Entries[1].ID != "0192-b" {
		t.Fatalf("tie broken wrong: %s then %s", page.Entries[0].ID, page.Entries[1].ID)
	}
}

func TestTimelinePaginationReturnsEveryEventOnce(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	own := []entity.IncidentEvent{
		{ID: "i1", At: at(0), Kind: entity.IncidentEventDeclared},
		{ID: "i2", At: at(3), Kind: entity.IncidentEventStatusChanged},
		{ID: "i3", At: at(5), Kind: entity.IncidentEventResolved},
	}
	linked := []entity.AlertEvent{
		{ID: "a1", AlertID: "al-1", At: at(1), Kind: entity.AlertEventNotified},
		{ID: "a2", AlertID: "al-1", At: at(4), Kind: entity.AlertEventAcked},
	}
	inc := entity.Incident{ID: "inc-1", WorkspaceID: "ws-1", Alerts: []entity.IncidentAlert{{AlertID: "al-1"}}}
	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").Return(inc, nil).AnyTimes()
	h.incidents.EXPECT().ListEvents(gomock.Any(), "ws-1", "inc-1", gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _, _ string, _ []entity.IncidentEventCategory, after entity.TimelineCursor, limit int) ([]entity.IncidentEvent, error) {
			out := []entity.IncidentEvent{}
			for _, e := range own {
				if after.Zero() || after.Before(entity.TimelineCursor{At: e.At, ID: e.ID}) {
					out = append(out, e)
				}
				if len(out) == limit {
					break
				}
			}
			return out, nil
		}).AnyTimes()
	h.alerts.EXPECT().ListEventsForAlerts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ []string, _ []entity.AlertEventKind, after entity.TimelineCursor, limit int) ([]entity.AlertEvent, error) {
			out := []entity.AlertEvent{}
			for _, e := range linked {
				if after.Zero() || after.Before(entity.TimelineCursor{At: e.At, ID: e.ID}) {
					out = append(out, e)
				}
				if len(out) == limit {
					break
				}
			}
			return out, nil
		}).AnyTimes()

	seen := []string{}
	cursor := ""
	for pages := 0; pages < len(own)+len(linked)+1; pages++ {
		page, err := h.srv.ListTimeline(adminCtx(), "acme", "inc-1", entity.TimelineFilter{Cursor: cursor, Limit: 2})
		if err != nil {
			t.Fatalf("list timeline: %v", err)
		}
		for _, e := range page.Entries {
			seen = append(seen, e.ID)
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	want := []string{"i1", "a1", "i2", "a2", "i3"}
	if len(seen) != len(want) {
		t.Fatalf("paged entries = %v, want %v", seen, want)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("paged entries = %v, want %v", seen, want)
		}
	}
}

func TestAddImageAttachmentWithoutStorage(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
	}, nil)
	h.incidents.EXPECT().CountAttachments(gomock.Any(), "ws-1", "ev-1").Return(0, nil)
	h.blobs.EXPECT().Enabled(gomock.Any()).Return(false)

	_, err := h.srv.AddAttachment(adminCtx(), "acme", "inc-1", "ev-1", entity.NewAttachment{
		Kind: entity.AttachmentImage, Label: "dashboard.png", ContentType: "image/png", SizeBytes: 2048,
	}, nil)
	if !errors.Is(err, entity.ErrAttachmentStorageUnavailable) {
		t.Fatalf("err = %v, want ErrAttachmentStorageUnavailable", err)
	}
}

func TestAddLinkAttachmentWorksWithoutStorage(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
	}, nil)
	h.incidents.EXPECT().CountAttachments(gomock.Any(), "ws-1", "ev-1").Return(0, nil)
	h.incidents.EXPECT().AddAttachment(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in entity.IncidentEventAttachment) (entity.IncidentEventAttachment, error) {
			in.ID = "att-1"
			return in, nil
		})
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	out, err := h.srv.AddAttachment(adminCtx(), "acme", "inc-1", "ev-1", entity.NewAttachment{
		Kind: entity.AttachmentLink, Label: "Grafana panel", URL: "https://grafana.example.com/d/abc",
	}, nil)
	if err != nil {
		t.Fatalf("add link attachment: %v", err)
	}
	if out.ID != "att-1" || out.ObjectKey != "" {
		t.Errorf("attachment = %+v, want a stored link with no object key", out)
	}
}

func TestAttachmentLimitPerEntry(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
	}, nil)
	h.incidents.EXPECT().CountAttachments(gomock.Any(), "ws-1", "ev-1").Return(entity.AttachmentsPerEntryMax, nil)

	_, err := h.srv.AddAttachment(adminCtx(), "acme", "inc-1", "ev-1", entity.NewAttachment{
		Kind: entity.AttachmentLog, Label: "app.log", Body: "panic: nil map",
	}, nil)
	if !errors.Is(err, entity.ErrAttachmentsPerEntryExceeded) {
		t.Fatalf("err = %v, want ErrAttachmentsPerEntryExceeded", err)
	}
}
