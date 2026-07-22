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

func alertProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to alerts in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid request", validationDetail(err), "")
	case errors.Is(err, entity.ErrAlertBulkEmpty):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Nothing selected", "Select at least one alert.", "")
	case errors.Is(err, entity.ErrSilenceWindow):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid window", "A silence has to end after it starts.", "")
	case errors.Is(err, entity.ErrSilenceEnded):
		return http.StatusConflict, prob(http.StatusConflict, "Already ended", "That silence has already ended.", "")
	case errors.Is(err, entity.ErrAlertNotFound), errors.Is(err, entity.ErrSilenceNotFound),
		errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That alert does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func (h *handler) alertDTO(a entity.Alert) api.Alert {
	labels := map[string]string{}
	for k, v := range a.Labels {
		labels[k] = v
	}
	links := make([]api.AlertLink, 0, len(a.Links))
	for _, l := range a.Links {
		links = append(links, api.AlertLink{Kind: api.AlertLinkKind(l.Kind), Label: l.Label, Url: l.URL})
	}
	timeline := make([]api.AlertEvent, 0, len(a.Timeline))
	for _, e := range a.Timeline {
		timeline = append(timeline, api.AlertEvent{
			Id: e.ID, At: e.At, Kind: string(e.Kind), Text: e.Text, Result: e.Result,
		})
	}

	dto := api.Alert{
		Id:              a.ID,
		DedupKey:        a.DedupKey,
		Title:           a.Title,
		Description:     a.Description,
		Severity:        api.AlertSeverity(a.Severity),
		Status:          api.AlertStatus(a.Status),
		Source:          firstNonBlank(a.SourceSlug, a.SourceLabel),
		Service:         a.ServiceName,
		Labels:          labels,
		Count:           a.Count,
		StartedAt:       a.StartedAt,
		LastSeenAt:      a.LastSeenAt,
		RoutedPolicyRef: a.RoutedPolicyRef,
		Suppressed:      a.SuppressedBySilenceID != "",
		Payload:         a.Payload,
		Links:           links,
		Timeline:        timeline,
	}
	if a.GroupKey != "" {
		dto.GroupKey = &a.GroupKey
	}
	if !a.AckedAt.IsZero() {
		at := a.AckedAt
		dto.AcknowledgedAt = &at
	}
	if !a.ResolvedAt.IsZero() {
		at := a.ResolvedAt
		dto.ResolvedAt = &at
	}
	if a.AckedByLabel != "" {
		label := a.AckedByLabel
		dto.AckedBy = &label
	}
	if a.ResolveMode != "" {
		mode := string(a.ResolveMode)
		dto.ResolveMode = &mode
	}
	return dto
}

func firstNonBlank(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func (h *handler) ListAlerts(ctx context.Context, request api.ListAlertsRequestObject) (api.ListAlertsResponseObject, error) {
	filter := entity.AlertFilter{}
	if request.Params.Status != nil {
		for _, s := range *request.Params.Status {
			filter.Statuses = append(filter.Statuses, entity.AlertStatus(s))
		}
	}
	if request.Params.Severity != nil {
		for _, s := range *request.Params.Severity {
			filter.Severities = append(filter.Severities, entity.AlertSeverity(s))
		}
	}
	if request.Params.Query != nil {
		filter.Query = *request.Params.Query
	}
	if request.Params.Cursor != nil {
		filter.Cursor = *request.Params.Cursor
	}
	if request.Params.Limit != nil {
		filter.Limit = *request.Params.Limit
	}

	list, next, err := h.alerts.List(ctx, request.WorkspaceId, filter)
	if err != nil {
		status, p := alertProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListAlerts401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListAlerts403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListAlerts404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}

	items := make([]api.Alert, 0, len(list))
	for _, a := range list {
		items = append(items, h.alertDTO(a))
	}
	out := api.ListAlerts200JSONResponse{Items: items}
	if next != "" {
		out.NextCursor = &next
	}
	return out, nil
}

func (h *handler) GetAlert(ctx context.Context, request api.GetAlertRequestObject) (api.GetAlertResponseObject, error) {
	alert, err := h.alerts.Get(ctx, request.WorkspaceId, request.AlertId)
	if err != nil {
		status, p := alertProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetAlert401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetAlert403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetAlert404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.GetAlert200JSONResponse(h.alertDTO(alert)), nil
}

func (h *handler) UpdateAlertStatus(ctx context.Context, request api.UpdateAlertStatusRequestObject) (api.UpdateAlertStatusResponseObject, error) {
	var (
		updated int
		err     error
	)
	if request.Body.Status == api.AlertStatusRequestStatusResolved {
		updated, err = h.alerts.Resolve(ctx, request.WorkspaceId, request.Body.Ids)
	} else {
		updated, err = h.alerts.Acknowledge(ctx, request.WorkspaceId, request.Body.Ids)
	}
	if err != nil {
		status, p := alertProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateAlertStatus400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateAlertStatus401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateAlertStatus403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateAlertStatus404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateAlertStatus200JSONResponse{Updated: updated}, nil
}

func (h *handler) ListIngestFailures(ctx context.Context, request api.ListIngestFailuresRequestObject) (api.ListIngestFailuresResponseObject, error) {
	limit := 0
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	failures, err := h.alerts.Failures(ctx, request.WorkspaceId, limit)
	if err != nil {
		status, p := alertProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListIngestFailures401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListIngestFailures403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListIngestFailures404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.IngestFailure, 0, len(failures))
	for _, f := range failures {
		items = append(items, api.IngestFailure{
			Id:      f.ID,
			Source:  f.SourceSlug,
			Reason:  string(f.Reason),
			Detail:  f.Detail,
			Payload: f.Payload,
			At:      f.At,
		})
	}
	return api.ListIngestFailures200JSONResponse{Items: items}, nil
}

func silenceDTO(s entity.Silence, now time.Time) api.Silence {
	conds := make([]api.SilenceCondition, 0, len(s.Conditions))
	for _, c := range s.Conditions {
		conds = append(conds, api.SilenceCondition{Field: api.SilenceConditionField(c.Field), Value: c.Value})
	}
	return api.Silence{
		Id:         s.ID,
		Kind:       api.SilenceKind(s.Kind),
		State:      api.SilenceState(s.State(now)),
		Reason:     s.Reason,
		CreatedBy:  s.CreatedBy,
		Conditions: conds,
		StartsAt:   s.StartsAt,
		EndsAt:     s.EndsAt,
	}
}

func (h *handler) ListSilences(ctx context.Context, request api.ListSilencesRequestObject) (api.ListSilencesResponseObject, error) {
	list, err := h.silences.List(ctx, request.WorkspaceId)
	if err != nil {
		status, p := alertProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListSilences401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListSilences403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListSilences404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	now := time.Now().UTC()
	items := make([]api.Silence, 0, len(list))
	for _, s := range list {
		items = append(items, silenceDTO(s, now))
	}
	return api.ListSilences200JSONResponse{Items: items}, nil
}

func (h *handler) CreateSilence(ctx context.Context, request api.CreateSilenceRequestObject) (api.CreateSilenceResponseObject, error) {
	in := entity.NewSilence{
		Kind:     entity.SilenceKindSilence,
		StartsAt: request.Body.StartsAt,
		EndsAt:   request.Body.EndsAt,
	}
	if request.Body.Kind != nil {
		in.Kind = entity.SilenceKind(*request.Body.Kind)
	}
	if request.Body.Reason != nil {
		in.Reason = *request.Body.Reason
	}
	for _, c := range request.Body.Conditions {
		in.Conditions = append(in.Conditions, entity.SilenceCondition{Field: string(c.Field), Value: c.Value})
	}

	created, err := h.silences.Create(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := alertProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.CreateSilence400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.CreateSilence401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.CreateSilence403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.CreateSilence404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.CreateSilence201JSONResponse(silenceDTO(created, time.Now().UTC())), nil
}

func (h *handler) EndSilence(ctx context.Context, request api.EndSilenceRequestObject) (api.EndSilenceResponseObject, error) {
	if err := h.silences.End(ctx, request.WorkspaceId, request.SilenceId); err != nil {
		status, p := alertProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.EndSilence401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.EndSilence403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.EndSilence404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.EndSilence409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.EndSilence204Response{}, nil
}
