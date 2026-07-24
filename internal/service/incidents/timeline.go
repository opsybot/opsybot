package incidents

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

func (s *srv) ListTimeline(ctx context.Context, workspaceSlug, id string, filter entity.TimelineFilter) (entity.TimelinePage, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.TimelinePage{}, err
	}
	inc, err := s.incidents.GetByID(ctx, ws.ID, id)
	if err != nil {
		return entity.TimelinePage{}, err
	}
	return s.timeline(ctx, inc, filter)
}

func (s *srv) timeline(ctx context.Context, inc entity.Incident, filter entity.TimelineFilter) (entity.TimelinePage, error) {
	limit := filter.Limit
	if limit <= 0 || limit > entity.TimelineMaxPageSize {
		limit = entity.TimelineDefaultPageSize
	}
	after, err := entity.ParseTimelineCursor(filter.Cursor)
	if err != nil {
		return entity.TimelinePage{}, err
	}

	own, err := s.incidents.ListEvents(ctx, inc.WorkspaceID, inc.ID, filter.Categories, after, limit+1)
	if err != nil {
		return entity.TimelinePage{}, err
	}

	titles := make(map[string]string, len(inc.Alerts))
	alertIDs := make([]string, 0, len(inc.Alerts))
	for _, a := range inc.Alerts {
		alertIDs = append(alertIDs, a.AlertID)
		titles[a.AlertID] = a.Title
	}
	kinds := entity.AlertEventKindsForCategories(filter.Categories)
	var linked []entity.AlertEvent
	if len(filter.Categories) == 0 || len(kinds) > 0 {
		linked, err = s.alerts.ListEventsForAlerts(ctx, alertIDs, kinds, after, limit+1)
		if err != nil {
			return entity.TimelinePage{}, err
		}
	}

	merged := make([]entity.IncidentEvent, 0, limit+1)
	i, j := 0, 0
	for len(merged) <= limit && (i < len(own) || j < len(linked)) {
		takeOwn := j >= len(linked)
		if !takeOwn && i < len(own) {
			takeOwn = entity.TimelineCursor{At: own[i].At, ID: own[i].ID}.
				Before(entity.TimelineCursor{At: linked[j].At, ID: linked[j].ID})
		}
		if takeOwn {
			merged = append(merged, own[i])
			i++
			continue
		}
		merged = append(merged, alertEntry(inc, linked[j], titles[linked[j].AlertID]))
		j++
	}

	page := entity.TimelinePage{}
	if len(merged) > limit {
		merged = merged[:limit]
		last := merged[limit-1]
		page.NextCursor = entity.TimelineCursor{At: last.At, ID: last.ID}.Encode()
	}
	page.Entries = merged
	return page, nil
}

func alertEntry(inc entity.Incident, event entity.AlertEvent, title string) entity.IncidentEvent {
	return entity.IncidentEvent{
		ID:          event.ID,
		IncidentID:  inc.ID,
		WorkspaceID: inc.WorkspaceID,
		At:          event.At,
		Category:    entity.CategoryForAlertEvent(event.Kind),
		Source:      entity.IncidentSourceSystem,
		Text:        event.Text,
		AlertID:     event.AlertID,
		AlertTitle:  title,
		AlertKind:   event.Kind,
		Result:      entity.AlertEventOutcome(event.Kind, event.Result),
	}
}

func (s *srv) AddTimelineEntry(ctx context.Context, workspaceSlug, id string, in entity.NewTimelineEntry) (entity.IncidentEvent, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.IncidentEvent{}, err
	}
	in.Text = strings.TrimSpace(in.Text)
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	if in.Category == "" {
		in.Category = entity.IncidentCategoryObservation
	}
	if !in.Source.Manual() {
		in.Source = entity.IncidentSourceUI
	}
	if err := in.Validate(); err != nil {
		return entity.IncidentEvent{}, err
	}
	now := time.Now().UTC()
	at, retroactive, err := in.Resolve(now)
	if err != nil {
		return entity.IncidentEvent{}, err
	}

	var created entity.IncidentEvent
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		inc, err := s.incidents.GetByID(ctx, ws.ID, id)
		if err != nil {
			return err
		}
		created, err = s.incidents.AppendEvent(ctx, entity.IncidentEvent{
			IncidentID:     inc.ID,
			WorkspaceID:    ws.ID,
			At:             at,
			Kind:           entity.IncidentEventNote,
			Category:       in.Category,
			Source:         in.Source,
			Text:           in.Text,
			Actor:          actor.Label,
			ActorUserID:    actor.UserID,
			Retroactive:    retroactive,
			IdempotencyKey: in.IdempotencyKey,
		})
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.IncidentEvent{}, err
	}
	return created, nil
}

func (s *srv) EditTimelineEntry(ctx context.Context, workspaceSlug, id, entryID string, in entity.TimelineEdit) (entity.IncidentEvent, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.IncidentEvent{}, err
	}
	in.Text = strings.TrimSpace(in.Text)
	if err := in.Validate(); err != nil {
		return entity.IncidentEvent{}, err
	}
	now := time.Now().UTC()

	var updated entity.IncidentEvent
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		entry, err := s.entry(ctx, ws.ID, id, entryID)
		if err != nil {
			return err
		}
		if entry.Kind != entity.IncidentEventNote {
			return entity.ErrTimelineEntryNotEditable
		}
		if entry.Text == in.Text && entry.Category == in.Category {
			updated = entry
			return nil
		}
		if err := s.incidents.AppendRevision(ctx, entity.IncidentEventRevision{
			EventID:      entry.ID,
			WorkspaceID:  ws.ID,
			At:           now,
			EditorUserID: actor.UserID,
			EditorLabel:  actor.Label,
			Text:         entry.Text,
			Category:     entry.Category,
		}); err != nil {
			return err
		}
		if err := s.incidents.UpdateEvent(ctx, ws.ID, entry.ID, in, now, actor.UserID); err != nil {
			return err
		}
		updated, err = s.incidents.GetEvent(ctx, ws.ID, entry.ID)
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return entity.IncidentEvent{}, err
	}
	return updated, nil
}

func (s *srv) ListEntryRevisions(ctx context.Context, workspaceSlug, id, entryID string) ([]entity.IncidentEventRevision, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return nil, err
	}
	if _, err := s.entry(ctx, ws.ID, id, entryID); err != nil {
		return nil, err
	}
	return s.incidents.ListRevisions(ctx, ws.ID, entryID)
}

func (s *srv) AddAttachment(ctx context.Context, workspaceSlug, id, entryID string, in entity.NewAttachment, content io.Reader) (entity.IncidentEventAttachment, error) {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.IncidentEventAttachment{}, err
	}
	in.Label = strings.TrimSpace(in.Label)
	in.URL = strings.TrimSpace(in.URL)
	if err := in.Validate(); err != nil {
		return entity.IncidentEventAttachment{}, err
	}
	entry, err := s.entry(ctx, ws.ID, id, entryID)
	if err != nil {
		return entity.IncidentEventAttachment{}, err
	}
	count, err := s.incidents.CountAttachments(ctx, ws.ID, entry.ID)
	if err != nil {
		return entity.IncidentEventAttachment{}, err
	}
	if count >= entity.AttachmentsPerEntryMax {
		return entity.IncidentEventAttachment{}, entity.ErrAttachmentsPerEntryExceeded
	}

	attachment := entity.IncidentEventAttachment{
		EventID:     entry.ID,
		WorkspaceID: ws.ID,
		Kind:        in.Kind,
		Label:       in.Label,
		URL:         in.URL,
		Body:        in.Body,
		ContentType: in.ContentType,
		CreatedBy:   actor.UserID,
	}
	if in.Kind == entity.AttachmentImage {
		if !s.blobs.Enabled(ctx) {
			return entity.IncidentEventAttachment{}, entity.ErrAttachmentStorageUnavailable
		}
		if in.SizeBytes > entity.AttachmentUploadMaxBytes {
			return entity.IncidentEventAttachment{}, entity.ErrAttachmentTooLarge
		}
		key, err := entity.AttachmentObjectKey(ws.ID, entry.ID)
		if err != nil {
			return entity.IncidentEventAttachment{}, err
		}
		if err := s.blobs.Put(ctx, key, content, in.SizeBytes, in.ContentType); err != nil {
			return entity.IncidentEventAttachment{}, err
		}
		attachment.ObjectKey = key
		attachment.SizeBytes = in.SizeBytes
	}

	var stored entity.IncidentEventAttachment
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		stored, err = s.incidents.AddAttachment(ctx, attachment)
		if err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		if attachment.ObjectKey != "" {
			if removeErr := s.blobs.Remove(ctx, attachment.ObjectKey); removeErr != nil {
				return entity.IncidentEventAttachment{}, errors.Join(err, removeErr)
			}
		}
		return entity.IncidentEventAttachment{}, err
	}
	return stored, nil
}

func (s *srv) OpenAttachment(ctx context.Context, workspaceSlug, id, attachmentID string) (entity.IncidentEventAttachment, io.ReadCloser, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.IncidentEventAttachment{}, nil, err
	}
	attachment, err := s.incidents.GetAttachment(ctx, ws.ID, attachmentID)
	if err != nil {
		return entity.IncidentEventAttachment{}, nil, err
	}
	if _, err := s.entry(ctx, ws.ID, id, attachment.EventID); err != nil {
		return entity.IncidentEventAttachment{}, nil, entity.ErrAttachmentNotFound
	}
	if attachment.ObjectKey == "" {
		return entity.IncidentEventAttachment{}, nil, entity.ErrAttachmentNotFound
	}
	body, err := s.blobs.Open(ctx, attachment.ObjectKey)
	if err != nil {
		return entity.IncidentEventAttachment{}, nil, err
	}
	return attachment, body, nil
}

func (s *srv) RemoveAttachment(ctx context.Context, workspaceSlug, id, attachmentID string) error {
	actor, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionWrite, entity.PolicyObjectIncidents)
	if err != nil {
		return err
	}
	attachment, err := s.incidents.GetAttachment(ctx, ws.ID, attachmentID)
	if err != nil {
		return err
	}
	entry, err := s.entry(ctx, ws.ID, id, attachment.EventID)
	if err != nil {
		return entity.ErrAttachmentNotFound
	}
	if entry.Kind != entity.IncidentEventNote {
		return entity.ErrTimelineEntryNotEditable
	}
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.incidents.RemoveAttachment(ctx, ws.ID, attachmentID); err != nil {
			return err
		}
		return s.audit.Create(ctx, s.event(actor, ws.ID, entity.ActionIncidentUpdated, id))
	})
	if err != nil {
		return err
	}
	if attachment.ObjectKey == "" {
		return nil
	}
	if err := s.blobs.Remove(ctx, attachment.ObjectKey); err != nil && !errors.Is(err, entity.ErrAttachmentStorageUnavailable) {
		return err
	}
	return nil
}

func (s *srv) ExportTimeline(ctx context.Context, workspaceSlug, id string) (entity.TimelineExport, error) {
	_, ws, err := s.authorize(ctx, workspaceSlug, entity.PolicyActionRead, entity.PolicyObjectIncidents)
	if err != nil {
		return entity.TimelineExport{}, err
	}
	inc, err := s.hydrate(ctx, ws.ID, id)
	if err != nil {
		return entity.TimelineExport{}, err
	}
	entries := make([]entity.IncidentEvent, 0, entity.TimelineDefaultPageSize)
	cursor := ""
	for len(entries) < entity.TimelineExportMaxEntries {
		page, err := s.timeline(ctx, inc, entity.TimelineFilter{Cursor: cursor, Limit: entity.TimelineMaxPageSize})
		if err != nil {
			return entity.TimelineExport{}, err
		}
		entries = append(entries, page.Entries...)
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	return entity.TimelineExport{Incident: inc, Entries: entries, ExportedAt: time.Now().UTC()}, nil
}

func (s *srv) entry(ctx context.Context, workspaceID, incidentID, entryID string) (entity.IncidentEvent, error) {
	entry, err := s.incidents.GetEvent(ctx, workspaceID, entryID)
	if err != nil {
		return entity.IncidentEvent{}, err
	}
	if entry.IncidentID != incidentID {
		return entity.IncidentEvent{}, entity.ErrTimelineEntryNotFound
	}
	return entry, nil
}
