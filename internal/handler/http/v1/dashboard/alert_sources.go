package dashboard

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func sourceProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to alert sources in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid source", validationDetail(err), "")
	case errors.Is(err, entity.ErrAlertSourceMappingEmpty):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Mapping required", "A generic source needs a field mapping.", "")
	case errors.Is(err, entity.ErrAlertMonitorFormat):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Use a monitor", "Heartbeats are created as monitors, which carry the check-in interval.", "")
	case errors.Is(err, entity.ErrAlertSourceSlugTaken):
		return http.StatusConflict, prob(http.StatusConflict, "Name taken", "A source already goes by that name.", "")
	case errors.Is(err, entity.ErrAlertSourceNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That source does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func (h *handler) sourceDTO(s entity.AlertSource) api.AlertSource {
	mapping := make([]api.AlertSourceMapping, 0, len(s.Mapping))
	for _, m := range s.Mapping {
		mapping = append(mapping, api.AlertSourceMapping{Field: m.Field, Path: m.Path})
	}
	status := api.AlertSourceStatusActive
	if s.Paused {
		status = api.AlertSourceStatusPaused
	}
	dto := api.AlertSource{
		Id:               s.ID,
		Slug:             s.Slug,
		Name:             s.Name,
		Format:           api.AlertSourceFormat(s.Format),
		Status:           status,
		Health:           api.AlertSourceHealth(s.Health(time.Now().UTC())),
		IngestUrl:        h.ingestURL(s.IngestToken),
		RequireSignature: s.RequireSignature,
		DefaultSeverity:  api.AlertSourceDefaultSeverity(s.DefaultSeverity),
		FailureCount:     s.FailureCount,
		Mapping:          mapping,
	}
	if s.SigningSecret != "" {
		secret := s.SigningSecret
		dto.SigningSecret = &secret
	}
	if !s.SecretRotatedAt.IsZero() {
		at := s.SecretRotatedAt
		dto.SecretRotatedAt = &at
	}
	if !s.LastEventAt.IsZero() {
		at := s.LastEventAt
		dto.LastEventAt = &at
	}
	return dto
}

func (h *handler) ingestURL(token string) string {
	base := strings.TrimRight(h.ingestBaseURL, "/")
	if base == "" {
		base = strings.TrimRight(h.cfg.BaseURL, "/")
	}
	return base + "/v1/ingest/e/" + token
}

func (h *handler) GetAlertSourceVolume(ctx context.Context, request api.GetAlertSourceVolumeRequestObject) (api.GetAlertSourceVolumeResponseObject, error) {
	volume, err := h.sources.Volume(ctx, request.WorkspaceId)
	if err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetAlertSourceVolume401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetAlertSourceVolume403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetAlertSourceVolume404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.GetAlertSourceVolume200JSONResponse{
		WindowHours: int(entity.SourceVolumeWindow / time.Hour),
		Sources:     volume,
	}, nil
}

func (h *handler) ListAlertSources(ctx context.Context, request api.ListAlertSourcesRequestObject) (api.ListAlertSourcesResponseObject, error) {
	list, err := h.sources.List(ctx, request.WorkspaceId)
	if err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListAlertSources401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListAlertSources403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListAlertSources404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.AlertSource, 0, len(list))
	for _, s := range list {
		items = append(items, h.sourceDTO(s))
	}
	return api.ListAlertSources200JSONResponse{Items: items}, nil
}

func (h *handler) GetAlertSource(ctx context.Context, request api.GetAlertSourceRequestObject) (api.GetAlertSourceResponseObject, error) {
	src, err := h.sources.Get(ctx, request.WorkspaceId, request.SourceSlug)
	if err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetAlertSource401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetAlertSource403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetAlertSource404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.GetAlertSource200JSONResponse(h.sourceDTO(src)), nil
}

func (h *handler) CreateAlertSource(ctx context.Context, request api.CreateAlertSourceRequestObject) (api.CreateAlertSourceResponseObject, error) {
	in := entity.NewAlertSource{
		Name:            request.Body.Name,
		Format:          entity.SourceFormat(request.Body.Format),
		DefaultSeverity: entity.SeverityWarning,
	}
	if request.Body.Slug != nil && strings.TrimSpace(*request.Body.Slug) != "" {
		in.Slug = *request.Body.Slug
	} else {
		in.Slug = entity.Slugify(request.Body.Name)
	}
	if request.Body.DefaultSeverity != nil {
		in.DefaultSeverity = entity.AlertSeverity(*request.Body.DefaultSeverity)
	}
	if request.Body.RequireSignature != nil {
		in.RequireSignature = *request.Body.RequireSignature
	}

	created, err := h.sources.Create(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.CreateAlertSource400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.CreateAlertSource401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.CreateAlertSource403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.CreateAlertSource404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.CreateAlertSource409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.CreateAlertSource201JSONResponse(h.sourceDTO(created)), nil
}

func (h *handler) DeleteAlertSource(ctx context.Context, request api.DeleteAlertSourceRequestObject) (api.DeleteAlertSourceResponseObject, error) {
	if err := h.sources.Delete(ctx, request.WorkspaceId, request.SourceSlug); err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DeleteAlertSource401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeleteAlertSource403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeleteAlertSource404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeleteAlertSource204Response{}, nil
}

func (h *handler) PauseAlertSource(ctx context.Context, request api.PauseAlertSourceRequestObject) (api.PauseAlertSourceResponseObject, error) {
	if err := h.sources.SetPaused(ctx, request.WorkspaceId, request.SourceSlug, true); err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.PauseAlertSource401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.PauseAlertSource403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.PauseAlertSource404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.PauseAlertSource204Response{}, nil
}

func (h *handler) ResumeAlertSource(ctx context.Context, request api.ResumeAlertSourceRequestObject) (api.ResumeAlertSourceResponseObject, error) {
	if err := h.sources.SetPaused(ctx, request.WorkspaceId, request.SourceSlug, false); err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ResumeAlertSource401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ResumeAlertSource403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ResumeAlertSource404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ResumeAlertSource204Response{}, nil
}

func (h *handler) RotateAlertSourceSecret(ctx context.Context, request api.RotateAlertSourceSecretRequestObject) (api.RotateAlertSourceSecretResponseObject, error) {
	rotated, err := h.sources.RotateSecret(ctx, request.WorkspaceId, request.SourceSlug)
	if err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.RotateAlertSourceSecret401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.RotateAlertSourceSecret403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.RotateAlertSourceSecret404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.RotateAlertSourceSecret200JSONResponse(h.sourceDTO(rotated)), nil
}

func (h *handler) UpdateAlertSourceMapping(ctx context.Context, request api.UpdateAlertSourceMappingRequestObject) (api.UpdateAlertSourceMappingResponseObject, error) {
	mappings := make([]entity.SourceMapping, 0, len(request.Body.Mapping))
	for i, m := range request.Body.Mapping {
		mappings = append(mappings, entity.SourceMapping{Field: m.Field, Path: m.Path, Position: i})
	}
	saved, err := h.sources.SaveMapping(ctx, request.WorkspaceId, request.SourceSlug, mappings)
	if err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateAlertSourceMapping400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateAlertSourceMapping401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateAlertSourceMapping403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateAlertSourceMapping404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateAlertSourceMapping200JSONResponse(h.sourceDTO(saved)), nil
}

func (h *handler) ListAlertSourceEvents(ctx context.Context, request api.ListAlertSourceEventsRequestObject) (api.ListAlertSourceEventsResponseObject, error) {
	limit := 0
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	events, err := h.sources.Events(ctx, request.WorkspaceId, request.SourceSlug, limit)
	if err != nil {
		status, p := sourceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListAlertSourceEvents401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListAlertSourceEvents403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListAlertSourceEvents404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.AlertSourceEvent, 0, len(events))
	for _, e := range events {
		item := api.AlertSourceEvent{Id: e.ID, Outcome: string(e.Outcome), DedupKey: e.DedupKey, At: e.At}
		if e.AlertID != "" {
			id := e.AlertID
			item.AlertId = &id
		}
		items = append(items, item)
	}
	return api.ListAlertSourceEvents200JSONResponse{Items: items}, nil
}
