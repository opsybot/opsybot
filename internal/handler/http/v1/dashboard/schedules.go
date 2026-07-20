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

const dateLayout = "2006-01-02"

func (h *handler) feedURL(token string) string {
	return strings.TrimRight(h.cfg.BaseURL, "/") + "/v1/oncall/feed/" + token + ".ics"
}

func (h *handler) scheduleDTO(s entity.Schedule) api.Schedule {
	now := time.Now()
	dto := api.Schedule{
		Id:        s.ID,
		Slug:      s.Slug,
		Team:      s.TeamSlug,
		Timezone:  s.Timezone,
		Paused:    s.Paused,
		Archived:  s.Archived,
		FeedUrl:   h.feedURL(s.FeedToken),
		Layers:    layerDTOs(s.Layers),
		Overrides: overrideDTOs(s.Overrides),
		CreatedAt: s.CreatedAt,
	}
	if cover := s.OnCallAt(now); cover.UserID != "" {
		uid := cover.UserID
		dto.OnCallUserId = &uid
		if seg, ok := s.OnCallSegment(now); ok {
			end := seg.EndsAt
			dto.OnCallUntil = &end
		}
	}
	return dto
}

func layerDTOs(layers []entity.Layer) []api.ScheduleLayer {
	out := make([]api.ScheduleLayer, 0, len(layers))
	for _, l := range layers {
		out = append(out, api.ScheduleLayer{
			Id:           l.ID,
			Participants: l.Participants,
			Rotation:     api.ScheduleLayerRotation(l.Rotation),
			IntervalDays: l.IntervalDays,
			HandoverHour: l.HandoverHour,
			StartsOn:     l.StartsOn.Format(dateLayout),
			Restrictions: restrictionDTOs(l.Restrictions),
		})
	}
	return out
}

func restrictionDTOs(in []entity.Restriction) []api.ScheduleRestriction {
	out := make([]api.ScheduleRestriction, 0, len(in))
	for _, r := range in {
		out = append(out, api.ScheduleRestriction{Weekday: r.Weekday, StartMinute: r.StartMinute, EndMinute: r.EndMinute})
	}
	return out
}

func overrideDTOs(in []entity.Override) []api.ScheduleOverride {
	out := make([]api.ScheduleOverride, 0, len(in))
	for _, o := range in {
		out = append(out, overrideDTO(o))
	}
	return out
}

func overrideDTO(o entity.Override) api.ScheduleOverride {
	dto := api.ScheduleOverride{
		Id:        o.ID,
		UserId:    o.UserID,
		StartsAt:  o.StartsAt,
		EndsAt:    o.EndsAt,
		Reason:    o.Reason,
		CreatedAt: o.CreatedAt,
	}
	if o.CreatedByUserID != "" {
		by := o.CreatedByUserID
		dto.CreatedByUserId = &by
	}
	return dto
}

func segmentDTO(seg entity.Segment) api.Segment {
	dto := api.Segment{StartsAt: seg.StartsAt, EndsAt: seg.EndsAt, Override: seg.Override}
	if seg.UserID != "" {
		uid := seg.UserID
		dto.UserId = &uid
	}
	if seg.Via != "" {
		via := seg.Via
		dto.Via = &via
	}
	return dto
}

func segmentDTOs(in []entity.Segment) []api.Segment {
	out := make([]api.Segment, 0, len(in))
	for _, seg := range in {
		out = append(out, segmentDTO(seg))
	}
	return out
}

func calendarDTO(c entity.ScheduleCalendar) api.ScheduleCalendar {
	layers := make([]api.LayerCoverage, 0, len(c.Layers))
	for _, l := range c.Layers {
		layers = append(layers, api.LayerCoverage{Index: l.Index, Via: l.Via, Segments: segmentDTOs(l.Segments)})
	}
	handovers := make([]api.Handover, 0, len(c.Handovers))
	for _, ho := range c.Handovers {
		handovers = append(handovers, api.Handover{At: ho.At, FromUserId: ho.FromUserID, ToUserId: ho.ToUserID})
	}
	return api.ScheduleCalendar{
		Effective: segmentDTOs(c.Effective),
		Gaps:      segmentDTOs(c.Gaps),
		Handovers: handovers,
		Layers:    layers,
	}
}

func newLayers(in []api.ScheduleLayerInput) []entity.NewScheduleLayer {
	out := make([]entity.NewScheduleLayer, 0, len(in))
	for _, l := range in {
		out = append(out, entity.NewScheduleLayer{
			Participants: l.Participants,
			Rotation:     entity.Rotation(l.Rotation),
			IntervalDays: l.IntervalDays,
			HandoverHour: l.HandoverHour,
			StartsOn:     parseDate(l.StartsOn),
			Restrictions: newRestrictions(l.Restrictions),
		})
	}
	return out
}

func newRestrictions(in []api.ScheduleRestriction) []entity.Restriction {
	out := make([]entity.Restriction, 0, len(in))
	for _, r := range in {
		out = append(out, entity.Restriction{Weekday: r.Weekday, StartMinute: r.StartMinute, EndMinute: r.EndMinute})
	}
	return out
}

func parseDate(s string) time.Time {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func scheduleStateDetail(err error) string {
	switch {
	case errors.Is(err, entity.ErrScheduleArchived):
		return "This schedule is archived."
	case errors.Is(err, entity.ErrScheduleNotArchived):
		return "This schedule is not archived."
	case errors.Is(err, entity.ErrSchedulePaused):
		return "This schedule is already paused."
	default:
		return "This schedule is not paused."
	}
}

func scheduleProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to schedules in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid schedule", validationDetail(err), "")
	case errors.Is(err, entity.ErrScheduleTeamInvalid):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid team", "Pick a team that exists in this workspace.", "")
	case errors.Is(err, entity.ErrScheduleParticipant):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid participant", "Every participant must be an active member of this workspace.", "")
	case errors.Is(err, entity.ErrScheduleOverrideWindow):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid override", "The override has to end after it starts.", "")
	case errors.Is(err, entity.ErrScheduleOverrideConflict):
		return http.StatusConflict, prob(http.StatusConflict, "Override conflict", "That window overlaps an existing override.", "")
	case errors.Is(err, entity.ErrScheduleOverrideNoChange):
		return http.StatusConflict, prob(http.StatusConflict, "No change", "That person already holds this shift. Pick someone else, or change the window.", "")
	case errors.Is(err, entity.ErrScheduleSlugTaken):
		return http.StatusConflict, prob(http.StatusConflict, "Name taken", "A schedule already goes by that name.", "")
	case errors.Is(err, entity.ErrScheduleArchived), errors.Is(err, entity.ErrScheduleNotArchived),
		errors.Is(err, entity.ErrSchedulePaused), errors.Is(err, entity.ErrScheduleNotPaused):
		return http.StatusConflict, prob(http.StatusConflict, "Conflict", scheduleStateDetail(err), "")
	case errors.Is(err, entity.ErrScheduleNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That schedule does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func (h *handler) ListSchedules(ctx context.Context, request api.ListSchedulesRequestObject) (api.ListSchedulesResponseObject, error) {
	includeArchived := request.Params.IncludeArchived != nil && *request.Params.IncludeArchived
	scheds, err := h.schedules.List(ctx, request.WorkspaceId, includeArchived)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListSchedules401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListSchedules403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListSchedules404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.Schedule, 0, len(scheds))
	for _, s := range scheds {
		items = append(items, h.scheduleDTO(s))
	}
	return api.ListSchedules200JSONResponse{Items: items}, nil
}

func (h *handler) GetSchedule(ctx context.Context, request api.GetScheduleRequestObject) (api.GetScheduleResponseObject, error) {
	sched, err := h.schedules.Get(ctx, request.WorkspaceId, request.ScheduleSlug)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetSchedule404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.GetSchedule200JSONResponse(h.scheduleDTO(sched)), nil
}

func (h *handler) CreateSchedule(ctx context.Context, request api.CreateScheduleRequestObject) (api.CreateScheduleResponseObject, error) {
	if request.Body == nil {
		return api.CreateSchedule400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	sched, err := h.schedules.Create(ctx, request.WorkspaceId, entity.NewSchedule{
		Slug: request.Body.Name, TeamSlug: request.Body.Team, Timezone: request.Body.Timezone, Layers: newLayers(request.Body.Layers),
	})
	if err != nil {
		if resp := createScheduleError(err); resp != nil {
			return resp, nil
		}
		return nil, err
	}
	return api.CreateSchedule201JSONResponse(h.scheduleDTO(sched)), nil
}

func createScheduleError(err error) api.CreateScheduleResponseObject {
	status, p := scheduleProblem(err)
	switch status {
	case http.StatusBadRequest:
		return api.CreateSchedule400ApplicationProblemPlusJSONResponse(p)
	case http.StatusUnauthorized:
		return api.CreateSchedule401ApplicationProblemPlusJSONResponse(p)
	case http.StatusForbidden:
		return api.CreateSchedule403ApplicationProblemPlusJSONResponse(p)
	case http.StatusNotFound:
		return api.CreateSchedule404ApplicationProblemPlusJSONResponse(p)
	case http.StatusConflict:
		return api.CreateSchedule409ApplicationProblemPlusJSONResponse(p)
	}
	return nil
}

func (h *handler) UpdateSchedule(ctx context.Context, request api.UpdateScheduleRequestObject) (api.UpdateScheduleResponseObject, error) {
	if request.Body == nil {
		return api.UpdateSchedule400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	sched, err := h.schedules.Update(ctx, request.WorkspaceId, request.ScheduleSlug, entity.ScheduleUpdate{
		Slug: request.Body.Name, TeamSlug: request.Body.Team, Timezone: request.Body.Timezone, Layers: newLayers(request.Body.Layers),
	})
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateSchedule400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateSchedule404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.UpdateSchedule409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateSchedule200JSONResponse(h.scheduleDTO(sched)), nil
}

func (h *handler) DeleteSchedule(ctx context.Context, request api.DeleteScheduleRequestObject) (api.DeleteScheduleResponseObject, error) {
	err := h.schedules.Delete(ctx, request.WorkspaceId, request.ScheduleSlug)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DeleteSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeleteSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeleteSchedule404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.DeleteSchedule409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeleteSchedule204Response{}, nil
}

func (h *handler) ArchiveSchedule(ctx context.Context, request api.ArchiveScheduleRequestObject) (api.ArchiveScheduleResponseObject, error) {
	sched, err := h.schedules.Archive(ctx, request.WorkspaceId, request.ScheduleSlug)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ArchiveSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ArchiveSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ArchiveSchedule404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.ArchiveSchedule409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ArchiveSchedule200JSONResponse(h.scheduleDTO(sched)), nil
}

func (h *handler) UnarchiveSchedule(ctx context.Context, request api.UnarchiveScheduleRequestObject) (api.UnarchiveScheduleResponseObject, error) {
	sched, err := h.schedules.Unarchive(ctx, request.WorkspaceId, request.ScheduleSlug)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.UnarchiveSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UnarchiveSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UnarchiveSchedule404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.UnarchiveSchedule409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UnarchiveSchedule200JSONResponse(h.scheduleDTO(sched)), nil
}

func (h *handler) PauseSchedule(ctx context.Context, request api.PauseScheduleRequestObject) (api.PauseScheduleResponseObject, error) {
	sched, err := h.schedules.Pause(ctx, request.WorkspaceId, request.ScheduleSlug)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.PauseSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.PauseSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.PauseSchedule404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.PauseSchedule409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.PauseSchedule200JSONResponse(h.scheduleDTO(sched)), nil
}

func (h *handler) ResumeSchedule(ctx context.Context, request api.ResumeScheduleRequestObject) (api.ResumeScheduleResponseObject, error) {
	sched, err := h.schedules.Resume(ctx, request.WorkspaceId, request.ScheduleSlug)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ResumeSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ResumeSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ResumeSchedule404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.ResumeSchedule409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ResumeSchedule200JSONResponse(h.scheduleDTO(sched)), nil
}

func (h *handler) DuplicateSchedule(ctx context.Context, request api.DuplicateScheduleRequestObject) (api.DuplicateScheduleResponseObject, error) {
	sched, err := h.schedules.Duplicate(ctx, request.WorkspaceId, request.ScheduleSlug)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DuplicateSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DuplicateSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DuplicateSchedule404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.DuplicateSchedule409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DuplicateSchedule201JSONResponse(h.scheduleDTO(sched)), nil
}

func (h *handler) AddScheduleOverride(ctx context.Context, request api.AddScheduleOverrideRequestObject) (api.AddScheduleOverrideResponseObject, error) {
	if request.Body == nil {
		return api.AddScheduleOverride400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	reason := ""
	if request.Body.Reason != nil {
		reason = *request.Body.Reason
	}
	override, err := h.schedules.AddOverride(ctx, request.WorkspaceId, request.ScheduleSlug, entity.NewOverride{
		UserID: request.Body.UserId, StartsAt: request.Body.StartsAt, EndsAt: request.Body.EndsAt, Reason: reason,
	})
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.AddScheduleOverride400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.AddScheduleOverride401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.AddScheduleOverride403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.AddScheduleOverride404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.AddScheduleOverride409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.AddScheduleOverride201JSONResponse(overrideDTO(override)), nil
}

func (h *handler) GetScheduleCalendar(ctx context.Context, request api.GetScheduleCalendarRequestObject) (api.GetScheduleCalendarResponseObject, error) {
	cal, err := h.schedules.Calendar(ctx, request.WorkspaceId, request.ScheduleSlug, request.Params.From, request.Params.To)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.GetScheduleCalendar400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.GetScheduleCalendar401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetScheduleCalendar403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetScheduleCalendar404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.GetScheduleCalendar200JSONResponse(calendarDTO(cal)), nil
}

func (h *handler) GetScheduleOnCall(ctx context.Context, request api.GetScheduleOnCallRequestObject) (api.GetScheduleOnCallResponseObject, error) {
	at := time.Now()
	if request.Params.At != nil {
		at = *request.Params.At
	}
	cover, until, err := h.schedules.OnCall(ctx, request.WorkspaceId, request.ScheduleSlug, at)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetScheduleOnCall401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetScheduleOnCall403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetScheduleOnCall404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	dto := api.ScheduleOnCall{Override: cover.Override}
	if cover.UserID != "" {
		uid := cover.UserID
		dto.UserId = &uid
	}
	if cover.Via != "" {
		via := cover.Via
		dto.Via = &via
	}
	if !until.IsZero() {
		dto.Until = &until
	}
	return api.GetScheduleOnCall200JSONResponse(dto), nil
}

func (h *handler) PreviewSchedule(ctx context.Context, request api.PreviewScheduleRequestObject) (api.PreviewScheduleResponseObject, error) {
	if request.Body == nil {
		return api.PreviewSchedule400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	cal, err := h.schedules.Preview(ctx, request.WorkspaceId, entity.NewSchedule{
		Timezone: request.Body.Timezone, Layers: newLayers(request.Body.Layers),
	}, request.Body.From, request.Body.To)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.PreviewSchedule400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.PreviewSchedule401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.PreviewSchedule403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.PreviewSchedule404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.PreviewSchedule200JSONResponse(calendarDTO(cal)), nil
}

func (h *handler) MyOnCall(ctx context.Context, request api.MyOnCallRequestObject) (api.MyOnCallResponseObject, error) {
	shifts, err := h.schedules.MyShifts(ctx, request.WorkspaceId, request.Params.From, request.Params.To)
	if err != nil {
		status, p := scheduleProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.MyOnCall401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.MyOnCall403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.MyOnCall404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.OnCallShift, 0, len(shifts))
	for _, sh := range shifts {
		items = append(items, api.OnCallShift{StartsAt: sh.StartsAt, EndsAt: sh.EndsAt, ScheduleSlug: sh.ScheduleSlug})
	}
	return api.MyOnCall200JSONResponse{Items: items}, nil
}
