package incidents

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
)

func testID(n int) string {
	return fmt.Sprintf("019f94e2-6957-7987-af9a-%012d", n)
}

func at(offset int) time.Time {
	return time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC).Add(time.Duration(offset) * time.Minute)
}

func TestEditRejectsAutomaticEntry(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetEventForUpdate(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
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

	h.incidents.EXPECT().GetEventForUpdate(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
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
		[]entity.IncidentEvent{{ID: testID(2), At: at(3), Kind: entity.IncidentEventNote, Text: "note"}}, nil)
	h.alerts.EXPECT().ListEventsForAlerts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]entity.AlertEvent{{ID: testID(1), AlertID: "al-1", At: at(3), Kind: entity.AlertEventAcked, Text: "acked"}}, nil)

	page, err := h.srv.ListTimeline(adminCtx(), "acme", "inc-1", entity.TimelineFilter{})
	if err != nil {
		t.Fatalf("list timeline: %v", err)
	}
	if page.Entries[0].ID != testID(1) || page.Entries[1].ID != testID(2) {
		t.Fatalf("tie broken wrong: %s then %s", page.Entries[0].ID, page.Entries[1].ID)
	}
}

func TestTimelinePaginationReturnsEveryEventOnce(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	own := []entity.IncidentEvent{
		{ID: testID(1), At: at(0), Kind: entity.IncidentEventDeclared},
		{ID: testID(3), At: at(3), Kind: entity.IncidentEventStatusChanged},
		{ID: testID(5), At: at(5), Kind: entity.IncidentEventResolved},
	}
	linked := []entity.AlertEvent{
		{ID: testID(2), AlertID: "al-1", At: at(1), Kind: entity.AlertEventNotified},
		{ID: testID(4), AlertID: "al-1", At: at(4), Kind: entity.AlertEventAcked},
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
	want := []string{testID(1), testID(2), testID(3), testID(4), testID(5)}
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
	h.incidents.EXPECT().GetEventForUpdate(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
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
	h.incidents.EXPECT().GetEventForUpdate(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
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

func TestRemoveAttachmentDeletesStoredObject(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetAttachment(gomock.Any(), "ws-1", "att-1").Return(entity.IncidentEventAttachment{
		ID: "att-1", EventID: "ev-1", Kind: entity.AttachmentImage, ObjectKey: "incidents/ws-1/ev-1/abc",
	}, nil)
	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
	}, nil)
	h.incidents.EXPECT().RemoveAttachment(gomock.Any(), "ws-1", "att-1").Return(nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	h.blobs.EXPECT().Remove(gomock.Any(), "incidents/ws-1/ev-1/abc").Return(nil)

	if err := h.srv.RemoveAttachment(adminCtx(), "acme", "inc-1", "att-1"); err != nil {
		t.Fatalf("remove attachment: %v", err)
	}
}

func TestRemoveAttachmentSucceedsWhenStorageIsGone(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetAttachment(gomock.Any(), "ws-1", "att-1").Return(entity.IncidentEventAttachment{
		ID: "att-1", EventID: "ev-1", Kind: entity.AttachmentImage, ObjectKey: "incidents/ws-1/ev-1/abc",
	}, nil)
	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
	}, nil)
	h.incidents.EXPECT().RemoveAttachment(gomock.Any(), "ws-1", "att-1").Return(nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	h.blobs.EXPECT().Remove(gomock.Any(), gomock.Any()).Return(entity.ErrAttachmentStorageUnavailable)

	if err := h.srv.RemoveAttachment(adminCtx(), "acme", "inc-1", "att-1"); err != nil {
		t.Fatalf("row removal must survive an unreachable object store: %v", err)
	}
}

func TestRemoveAttachmentRejectedOnAutomaticEntry(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetAttachment(gomock.Any(), "ws-1", "att-1").Return(entity.IncidentEventAttachment{
		ID: "att-1", EventID: "ev-1", Kind: entity.AttachmentLink,
	}, nil)
	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventDeclared,
	}, nil)

	err := h.srv.RemoveAttachment(adminCtx(), "acme", "inc-1", "att-1")
	if !errors.Is(err, entity.ErrTimelineEntryNotEditable) {
		t.Fatalf("err = %v, want ErrTimelineEntryNotEditable", err)
	}
}

func TestRemoveAttachmentFromAnotherIncidentIsNotFound(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetAttachment(gomock.Any(), "ws-1", "att-1").Return(entity.IncidentEventAttachment{
		ID: "att-1", EventID: "ev-9", Kind: entity.AttachmentLink,
	}, nil)
	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-9").Return(entity.IncidentEvent{
		ID: "ev-9", IncidentID: "inc-other", Kind: entity.IncidentEventNote,
	}, nil)

	err := h.srv.RemoveAttachment(adminCtx(), "acme", "inc-1", "att-1")
	if !errors.Is(err, entity.ErrAttachmentNotFound) {
		t.Fatalf("err = %v, want ErrAttachmentNotFound", err)
	}
}

func TestExportMarksTruncationWhenTheTimelineIsHuge(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()
	h.emptyDirectory()

	inc := entity.Incident{ID: "inc-1", WorkspaceID: "ws-1", Number: 4, Name: "Long incident"}
	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").Return(inc, nil)
	h.alerts.EXPECT().ListEventsForAlerts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()

	full := make([]entity.IncidentEvent, entity.TimelineMaxPageSize+1)
	for i := range full {
		full[i] = entity.IncidentEvent{
			ID:   testID(i),
			At:   at(i),
			Kind: entity.IncidentEventNote,
		}
	}
	h.incidents.EXPECT().ListEvents(gomock.Any(), "ws-1", "inc-1", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(full, nil).AnyTimes()

	export, err := h.srv.ExportTimeline(adminCtx(), "acme", "inc-1")
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !export.Truncated {
		t.Fatal("export past the cap must report truncation")
	}
	if !strings.Contains(export.Text(), "Truncated to the first") {
		t.Fatal("human-readable export must say it was truncated")
	}
}

func TestExportFollowsCursorPastTheRepositoryPageCap(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()
	h.emptyDirectory()

	total := entity.TimelineMaxPageSize * 2
	all := make([]entity.IncidentEvent, total)
	for i := range all {
		all[i] = entity.IncidentEvent{ID: testID(i), At: at(i), Kind: entity.IncidentEventNote}
	}

	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").
		Return(entity.Incident{ID: "inc-1", WorkspaceID: "ws-1", Number: 9, Name: "Long"}, nil)
	h.alerts.EXPECT().ListEventsForAlerts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()
	h.incidents.EXPECT().ListEvents(gomock.Any(), "ws-1", "inc-1", gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _, _ string, _ []entity.IncidentEventCategory, after entity.TimelineCursor, limit int) ([]entity.IncidentEvent, error) {
			if limit > entity.TimelineFetchLimit {
				limit = entity.TimelineFetchLimit
			}
			out := []entity.IncidentEvent{}
			for _, e := range all {
				if !after.Zero() && !after.Before(entity.TimelineCursor{At: e.At, ID: e.ID}) {
					continue
				}
				out = append(out, e)
				if len(out) == limit {
					break
				}
			}
			return out, nil
		}).AnyTimes()

	export, err := h.srv.ExportTimeline(adminCtx(), "acme", "inc-1")
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if len(export.Entries) != total {
		t.Fatalf("exported %d of %d entries", len(export.Entries), total)
	}
	if export.Truncated {
		t.Fatal("timeline is under the export cap and must not report truncation")
	}
}

func TestAddImageAttachmentRejectsScriptableTypes(t *testing.T) {
	blocked := []string{
		"image/svg+xml",
		"image/svg+xml; charset=utf-8",
		"IMAGE/SVG+XML",
		"text/html",
		"application/xhtml+xml",
		"",
	}
	for _, contentType := range blocked {
		t.Run(contentType, func(t *testing.T) {
			h := newHarness(t)
			h.authorizeOK()

			_, err := h.srv.AddAttachment(adminCtx(), "acme", "inc-1", "ev-1", entity.NewAttachment{
				Kind: entity.AttachmentImage, Label: "payload", ContentType: contentType, SizeBytes: 64,
			}, nil)
			if !entity.IsValidationError(err) {
				t.Fatalf("content type %q was accepted; err = %v", contentType, err)
			}
		})
	}
}

func TestAddImageAttachmentAcceptsRasterTypes(t *testing.T) {
	for _, contentType := range entity.AttachmentImageTypes {
		t.Run(contentType, func(t *testing.T) {
			h := newHarness(t)
			h.authorizeOK()

			h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
				ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
			}, nil)
			h.incidents.EXPECT().GetEventForUpdate(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
				ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
			}, nil)
			h.incidents.EXPECT().CountAttachments(gomock.Any(), "ws-1", "ev-1").Return(0, nil)
			h.blobs.EXPECT().Enabled(gomock.Any()).Return(true)
			h.blobs.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any(), int64(64), contentType).Return(nil)
			h.incidents.EXPECT().AddAttachment(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, in entity.IncidentEventAttachment) (entity.IncidentEventAttachment, error) {
					in.ID = "att-1"
					return in, nil
				})
			h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

			out, err := h.srv.AddAttachment(adminCtx(), "acme", "inc-1", "ev-1", entity.NewAttachment{
				Kind: entity.AttachmentImage, Label: "screenshot", ContentType: contentType, SizeBytes: 64,
			}, nil)
			if err != nil {
				t.Fatalf("content type %q rejected: %v", contentType, err)
			}
			if out.ObjectKey == "" {
				t.Fatal("stored image has no object key")
			}
		})
	}
}

func TestAddAttachmentRejectedOnAutomaticEntry(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventStatusChanged,
	}, nil)

	_, err := h.srv.AddAttachment(adminCtx(), "acme", "inc-1", "ev-1", entity.NewAttachment{
		Kind: entity.AttachmentLink, Label: "evidence", URL: "https://example.test/x",
	}, nil)
	if !errors.Is(err, entity.ErrTimelineEntryNotEditable) {
		t.Fatalf("err = %v, want ErrTimelineEntryNotEditable; an attachment on an automatic entry could never be removed", err)
	}
}

func TestRemoveAttachmentSurvivesObjectStoreFailure(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().GetAttachment(gomock.Any(), "ws-1", "att-1").Return(entity.IncidentEventAttachment{
		ID: "att-1", EventID: "ev-1", Kind: entity.AttachmentImage, ObjectKey: "incidents/ws-1/ev-1/abc",
	}, nil)
	h.incidents.EXPECT().GetEvent(gomock.Any(), "ws-1", "ev-1").Return(entity.IncidentEvent{
		ID: "ev-1", IncidentID: "inc-1", Kind: entity.IncidentEventNote,
	}, nil)
	h.incidents.EXPECT().RemoveAttachment(gomock.Any(), "ws-1", "att-1").Return(nil)
	h.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	h.blobs.EXPECT().Remove(gomock.Any(), gomock.Any()).Return(errors.New("bucket unreachable"))

	if err := h.srv.RemoveAttachment(adminCtx(), "acme", "inc-1", "att-1"); err != nil {
		t.Fatalf("the row is already deleted, so the caller must see success: %v", err)
	}
}

func TestAddAttachmentRequiresALabel(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	_, err := h.srv.AddAttachment(adminCtx(), "acme", "inc-1", "ev-1", entity.NewAttachment{
		Kind: entity.AttachmentLink, Label: "   ", URL: "https://example.test/x",
	}, nil)
	if !entity.IsValidationError(err) {
		t.Fatalf("err = %v, want a validation error for a blank label", err)
	}
}

func TestListTimelineRejectsMalformedCursor(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()
	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").
		Return(entity.Incident{ID: "inc-1", WorkspaceID: "ws-1"}, nil).AnyTimes()

	for _, cursor := range []string{
		"2026-07-24T10:00:00Z|abc",
		"2026-07-24T10:00:00Z|",
		"2026-07-24T10:00:00Z|'; DROP TABLE incident_events--",
		"not-a-time|019f94e2-6957-7987-af9a-00a25f112554",
		"missing-separator",
	} {
		t.Run(cursor, func(t *testing.T) {
			_, err := h.srv.ListTimeline(adminCtx(), "acme", "inc-1", entity.TimelineFilter{Cursor: cursor})
			if !errors.Is(err, entity.ErrTimelineCursorInvalid) {
				t.Fatalf("cursor %q gave err = %v, want ErrTimelineCursorInvalid (otherwise it reaches Postgres as a uuid comparand)", cursor, err)
			}
		})
	}
}

func TestListTimelineAcceptsAGeneratedCursor(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	inc := entity.Incident{ID: "inc-1", WorkspaceID: "ws-1"}
	h.incidents.EXPECT().GetByID(gomock.Any(), "ws-1", "inc-1").Return(inc, nil)
	h.incidents.EXPECT().ListEvents(gomock.Any(), "ws-1", "inc-1", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
	h.alerts.EXPECT().ListEventsForAlerts(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

	cursor := entity.TimelineCursor{At: at(0), ID: "019f94e2-6957-7987-af9a-00a25f112554"}.Encode()
	if _, err := h.srv.ListTimeline(adminCtx(), "acme", "inc-1", entity.TimelineFilter{Cursor: cursor}); err != nil {
		t.Fatalf("a cursor this service emitted was rejected: %v", err)
	}
}

func TestToggleFollowupRejectsAFollowupFromAnotherIncident(t *testing.T) {
	h := newHarness(t)
	h.authorizeOK()

	h.incidents.EXPECT().SetFollowupDone(gomock.Any(), "ws-1", "fu-1", true, gomock.Any()).Return(
		entity.IncidentFollowup{ID: "fu-1", IncidentID: "inc-other", Title: "Patch the indexer"}, nil)

	_, err := h.srv.ToggleFollowup(adminCtx(), "acme", "inc-1", "fu-1", true)
	if !errors.Is(err, entity.ErrFollowupNotFound) {
		t.Fatalf("err = %v, want ErrFollowupNotFound; the event would otherwise land on the wrong incident", err)
	}
}
