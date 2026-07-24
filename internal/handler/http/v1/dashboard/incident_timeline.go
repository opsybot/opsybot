package dashboard

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) ListIncidentTimeline(ctx context.Context, request api.ListIncidentTimelineRequestObject) (api.ListIncidentTimelineResponseObject, error) {
	filter := entity.TimelineFilter{}
	if request.Params.Cursor != nil {
		filter.Cursor = *request.Params.Cursor
	}
	if request.Params.Limit != nil {
		filter.Limit = *request.Params.Limit
	}
	if request.Params.Category != nil {
		for _, c := range *request.Params.Category {
			filter.Categories = append(filter.Categories, entity.IncidentEventCategory(c))
		}
	}
	page, err := h.incidents.ListTimeline(ctx, request.WorkspaceId, request.IncidentId, filter)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.ListIncidentTimeline400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.ListIncidentTimeline401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListIncidentTimeline403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListIncidentTimeline404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	out := api.TimelinePage{Entries: eventsDTO(page.Entries)}
	if page.NextCursor != "" {
		out.NextCursor = ptr(page.NextCursor)
	}
	return api.ListIncidentTimeline200JSONResponse(out), nil
}

func (h *handler) AddIncidentTimelineEntry(ctx context.Context, request api.AddIncidentTimelineEntryRequestObject) (api.AddIncidentTimelineEntryResponseObject, error) {
	if request.Body == nil {
		p := prob(http.StatusBadRequest, "Invalid entry", "Send a timeline entry to add.", "")
		return api.AddIncidentTimelineEntry400ApplicationProblemPlusJSONResponse(p), nil
	}
	in := entity.NewTimelineEntry{Text: request.Body.Text, Source: entity.IncidentSourceUI}
	if request.Body.Category != nil {
		in.Category = entity.IncidentEventCategory(*request.Body.Category)
	}
	if request.Body.At != nil {
		in.At = *request.Body.At
	}
	if request.Body.IdempotencyKey != nil {
		in.IdempotencyKey = *request.Body.IdempotencyKey
	}
	entry, err := h.incidents.AddTimelineEntry(ctx, request.WorkspaceId, request.IncidentId, in)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.AddIncidentTimelineEntry400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.AddIncidentTimelineEntry401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.AddIncidentTimelineEntry403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.AddIncidentTimelineEntry404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.AddIncidentTimelineEntry200JSONResponse(eventDTO(entry)), nil
}

func (h *handler) EditIncidentTimelineEntry(ctx context.Context, request api.EditIncidentTimelineEntryRequestObject) (api.EditIncidentTimelineEntryResponseObject, error) {
	if request.Body == nil {
		p := prob(http.StatusBadRequest, "Invalid entry", "Send the updated timeline entry.", "")
		return api.EditIncidentTimelineEntry400ApplicationProblemPlusJSONResponse(p), nil
	}
	in := entity.TimelineEdit{
		Text:     request.Body.Text,
		Category: entity.IncidentEventCategory(request.Body.Category),
	}
	entry, err := h.incidents.EditTimelineEntry(ctx, request.WorkspaceId, request.IncidentId, request.EntryId, in)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.EditIncidentTimelineEntry400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.EditIncidentTimelineEntry401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.EditIncidentTimelineEntry403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.EditIncidentTimelineEntry404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.EditIncidentTimelineEntry409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.EditIncidentTimelineEntry200JSONResponse(eventDTO(entry)), nil
}

func (h *handler) ListIncidentTimelineRevisions(ctx context.Context, request api.ListIncidentTimelineRevisionsRequestObject) (api.ListIncidentTimelineRevisionsResponseObject, error) {
	revisions, err := h.incidents.ListEntryRevisions(ctx, request.WorkspaceId, request.IncidentId, request.EntryId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListIncidentTimelineRevisions401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListIncidentTimelineRevisions403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListIncidentTimelineRevisions404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	out := api.TimelineRevisionList{Revisions: make([]api.TimelineRevision, 0, len(revisions))}
	for _, r := range revisions {
		item := api.TimelineRevision{
			Id:       r.ID,
			At:       r.At,
			Text:     r.Text,
			Category: api.TimelineCategory(r.Category),
		}
		if r.EditorLabel != "" {
			item.EditorLabel = ptr(r.EditorLabel)
		}
		out.Revisions = append(out.Revisions, item)
	}
	return api.ListIncidentTimelineRevisions200JSONResponse(out), nil
}

func (h *handler) AddIncidentTimelineAttachment(ctx context.Context, request api.AddIncidentTimelineAttachmentRequestObject) (api.AddIncidentTimelineAttachmentResponseObject, error) {
	var in entity.NewAttachment
	var content io.Reader

	switch {
	case request.JSONBody != nil:
		in = entity.NewAttachment{Kind: entity.AttachmentKind(request.JSONBody.Kind), Label: request.JSONBody.Label}
		if request.JSONBody.Url != nil {
			in.URL = *request.JSONBody.Url
		}
		if request.JSONBody.Body != nil {
			in.Body = *request.JSONBody.Body
		}
	case request.MultipartBody != nil:
		upload, body, err := readAttachmentUpload(request.MultipartBody)
		if err != nil {
			status, p := incidentProblem(err)
			if status == http.StatusRequestEntityTooLarge {
				return api.AddIncidentTimelineAttachment413ApplicationProblemPlusJSONResponse(p), nil
			}
			p = prob(http.StatusBadRequest, "Invalid upload", "Send a file and a label.", "")
			return api.AddIncidentTimelineAttachment400ApplicationProblemPlusJSONResponse(p), nil
		}
		in, content = upload, body
	default:
		p := prob(http.StatusBadRequest, "Invalid attachment", "Send an attachment to add.", "")
		return api.AddIncidentTimelineAttachment400ApplicationProblemPlusJSONResponse(p), nil
	}

	attachment, err := h.incidents.AddAttachment(ctx, request.WorkspaceId, request.IncidentId, request.EntryId, in, content)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.AddIncidentTimelineAttachment400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.AddIncidentTimelineAttachment401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.AddIncidentTimelineAttachment403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.AddIncidentTimelineAttachment404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.AddIncidentTimelineAttachment409ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusRequestEntityTooLarge:
			return api.AddIncidentTimelineAttachment413ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	dto := attachmentsDTO([]entity.IncidentEventAttachment{attachment})
	return api.AddIncidentTimelineAttachment200JSONResponse(dto[0]), nil
}

func (h *handler) RemoveIncidentTimelineAttachment(ctx context.Context, request api.RemoveIncidentTimelineAttachmentRequestObject) (api.RemoveIncidentTimelineAttachmentResponseObject, error) {
	err := h.incidents.RemoveAttachment(ctx, request.WorkspaceId, request.IncidentId, request.AttachmentId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.RemoveIncidentTimelineAttachment401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.RemoveIncidentTimelineAttachment403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.RemoveIncidentTimelineAttachment404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.RemoveIncidentTimelineAttachment409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.RemoveIncidentTimelineAttachment204Response{}, nil
}

func (h *handler) DownloadIncidentTimelineAttachment(ctx context.Context, request api.DownloadIncidentTimelineAttachmentRequestObject) (api.DownloadIncidentTimelineAttachmentResponseObject, error) {
	attachment, body, err := h.incidents.OpenAttachment(ctx, request.WorkspaceId, request.IncidentId, request.AttachmentId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DownloadIncidentTimelineAttachment401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DownloadIncidentTimelineAttachment403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DownloadIncidentTimelineAttachment404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.DownloadIncidentTimelineAttachment409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return attachmentContent{contentType: attachment.ContentType, size: attachment.SizeBytes, body: body}, nil
}

type attachmentContent struct {
	contentType string
	size        int64
	body        io.ReadCloser
}

func (a attachmentContent) VisitDownloadIncidentTimelineAttachmentResponse(w http.ResponseWriter) error {
	defer a.body.Close()
	contentType := a.contentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	if a.size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(a.size, 10))
	}
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, err := io.Copy(w, a.body)
	return err
}

func (h *handler) ExportIncidentTimeline(ctx context.Context, request api.ExportIncidentTimelineRequestObject) (api.ExportIncidentTimelineResponseObject, error) {
	export, err := h.incidents.ExportTimeline(ctx, request.WorkspaceId, request.IncidentId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ExportIncidentTimeline401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ExportIncidentTimeline403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ExportIncidentTimeline404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ExportIncidentTimeline200JSONResponse(api.TimelineExport{
		IncidentId: export.Incident.ID,
		Number:     export.Incident.Number,
		Name:       export.Incident.Name,
		ExportedAt: export.ExportedAt,
		Entries:    eventsDTO(export.Entries),
		Text:       export.Text(),
	}), nil
}

func readAttachmentUpload(reader *multipart.Reader) (entity.NewAttachment, io.Reader, error) {
	out := entity.NewAttachment{Kind: entity.AttachmentImage}
	var content bytes.Buffer
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return entity.NewAttachment{}, nil, err
		}
		switch part.FormName() {
		case "label":
			label, err := io.ReadAll(io.LimitReader(part, entity.AttachmentLabelMaxLength+1))
			if err != nil {
				return entity.NewAttachment{}, nil, err
			}
			out.Label = string(label)
		case "file":
			size, err := io.Copy(&content, io.LimitReader(part, entity.AttachmentUploadMaxBytes+1))
			if err != nil {
				return entity.NewAttachment{}, nil, err
			}
			if size > entity.AttachmentUploadMaxBytes {
				return entity.NewAttachment{}, nil, entity.ErrAttachmentTooLarge
			}
			out.SizeBytes = size
			out.ContentType = part.Header.Get("Content-Type")
			if out.Label == "" {
				out.Label = part.FileName()
			}
		}
		part.Close()
	}
	return out, &content, nil
}
