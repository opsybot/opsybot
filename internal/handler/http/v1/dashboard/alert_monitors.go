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

func monitorProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to heartbeat monitors in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid monitor", validationDetail(err), "")
	case errors.Is(err, entity.ErrAlertSourceSlugTaken):
		return http.StatusConflict, prob(http.StatusConflict, "Name taken", "A source already goes by that name.", "")
	case errors.Is(err, entity.ErrAlertMonitorNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That monitor does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func (h *handler) monitorDTO(m entity.AlertMonitor) api.AlertMonitor {
	now := time.Now().UTC()
	dueAt := m.DueAt()
	dto := api.AlertMonitor{
		Id:              m.ID,
		Slug:            m.Slug,
		Name:            m.Name,
		State:           api.AlertMonitorState(m.State(now)),
		IntervalSeconds: int(m.Interval / time.Second),
		GraceSeconds:    int(m.Grace / time.Second),
		PolicyRef:       m.PolicyRef,
		Severity:        api.AlertMonitorSeverity(m.Severity),
		CheckInUrl:      h.checkInURL(m.CheckInToken),
		Paused:          m.Paused,
		DueAt:           &dueAt,
	}
	if !m.LastCheckInAt.IsZero() {
		at := m.LastCheckInAt
		dto.LastCheckInAt = &at
	}
	return dto
}

func (h *handler) checkInURL(token string) string {
	base := strings.TrimRight(h.ingestBaseURL, "/")
	if base == "" {
		base = strings.TrimRight(h.cfg.BaseURL, "/")
	}
	return base + "/v1/ingest/hb/" + token
}

func (h *handler) ListAlertMonitors(ctx context.Context, request api.ListAlertMonitorsRequestObject) (api.ListAlertMonitorsResponseObject, error) {
	list, err := h.monitors.List(ctx, request.WorkspaceId)
	if err != nil {
		status, p := monitorProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListAlertMonitors401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListAlertMonitors403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListAlertMonitors404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.AlertMonitor, 0, len(list))
	for _, m := range list {
		items = append(items, h.monitorDTO(m))
	}
	return api.ListAlertMonitors200JSONResponse{Items: items}, nil
}

func (h *handler) GetAlertMonitor(ctx context.Context, request api.GetAlertMonitorRequestObject) (api.GetAlertMonitorResponseObject, error) {
	monitor, err := h.monitors.Get(ctx, request.WorkspaceId, request.MonitorId)
	if err != nil {
		status, p := monitorProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetAlertMonitor401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetAlertMonitor403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetAlertMonitor404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.GetAlertMonitor200JSONResponse(h.monitorDTO(monitor)), nil
}

func (h *handler) CreateAlertMonitor(ctx context.Context, request api.CreateAlertMonitorRequestObject) (api.CreateAlertMonitorResponseObject, error) {
	in := entity.NewAlertMonitor{
		Name:     request.Body.Name,
		Interval: time.Duration(request.Body.IntervalSeconds) * time.Second,
		Grace:    entity.MonitorGraceDefault,
	}
	if request.Body.Slug != nil {
		in.Slug = *request.Body.Slug
	}
	if request.Body.GraceSeconds != nil {
		in.Grace = time.Duration(*request.Body.GraceSeconds) * time.Second
	}
	if request.Body.PolicyRef != nil {
		in.PolicyRef = *request.Body.PolicyRef
	}
	if request.Body.Severity != nil {
		in.Severity = entity.AlertSeverity(*request.Body.Severity)
	}

	created, err := h.monitors.Create(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := monitorProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.CreateAlertMonitor400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.CreateAlertMonitor401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.CreateAlertMonitor403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.CreateAlertMonitor404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.CreateAlertMonitor409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.CreateAlertMonitor201JSONResponse(h.monitorDTO(created)), nil
}

func (h *handler) UpdateAlertMonitor(ctx context.Context, request api.UpdateAlertMonitorRequestObject) (api.UpdateAlertMonitorResponseObject, error) {
	in := entity.AlertMonitorUpdate{
		Name:     request.Body.Name,
		Interval: time.Duration(request.Body.IntervalSeconds) * time.Second,
		Grace:    entity.MonitorGraceDefault,
	}
	if request.Body.GraceSeconds != nil {
		in.Grace = time.Duration(*request.Body.GraceSeconds) * time.Second
	}
	if request.Body.PolicyRef != nil {
		in.PolicyRef = *request.Body.PolicyRef
	}
	if request.Body.Severity != nil {
		in.Severity = entity.AlertSeverity(*request.Body.Severity)
	}

	updated, err := h.monitors.Update(ctx, request.WorkspaceId, request.MonitorId, in)
	if err != nil {
		status, p := monitorProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateAlertMonitor400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateAlertMonitor401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateAlertMonitor403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateAlertMonitor404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateAlertMonitor200JSONResponse(h.monitorDTO(updated)), nil
}

func (h *handler) DeleteAlertMonitor(ctx context.Context, request api.DeleteAlertMonitorRequestObject) (api.DeleteAlertMonitorResponseObject, error) {
	if err := h.monitors.Delete(ctx, request.WorkspaceId, request.MonitorId); err != nil {
		status, p := monitorProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DeleteAlertMonitor401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeleteAlertMonitor403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeleteAlertMonitor404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeleteAlertMonitor204Response{}, nil
}
