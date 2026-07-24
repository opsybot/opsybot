package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func incidentProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to incidents in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid incident", validationDetail(err), "")
	case errors.Is(err, entity.ErrIncidentInvalidTransition):
		return http.StatusConflict, prob(http.StatusConflict, "Invalid transition", "That status change isn't allowed from where the incident is now.", "")
	case errors.Is(err, entity.ErrIncidentResolutionMissing):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Resolution needed", "Add a short resolution summary before resolving the incident.", "")
	case errors.Is(err, entity.ErrIncidentSeverityUnknown):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Unknown severity", "Pick a severity that exists in this workspace.", "")
	case errors.Is(err, entity.ErrIncidentSelfRelation):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid relation", "An incident can't be related to itself.", "")
	case errors.Is(err, entity.ErrIncidentFieldUnknown), errors.Is(err, entity.ErrIncidentFieldValueInvalid):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid custom field", "Check the custom field values against their definitions.", "")
	case errors.Is(err, entity.ErrIncidentLeadUnknown):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Unknown lead", "Pick an active member as the incident lead.", "")
	case errors.Is(err, entity.ErrServiceNotFound):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Unknown service", "One of the selected services no longer exists.", "")
	case errors.Is(err, entity.ErrFieldSlugTaken):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Duplicate field", "Two custom fields resolve to the same key. Rename one.", "")
	case errors.Is(err, entity.ErrTeamNotFound):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Unknown team", "Pick an existing team for this incident.", "")
	case errors.Is(err, entity.ErrIncidentNotFound), errors.Is(err, entity.ErrFollowupNotFound),
		errors.Is(err, entity.ErrAlertNotFound), errors.Is(err, entity.ErrWorkspaceNotFound),
		errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That incident does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func (h *handler) incidentDTO(inc entity.Incident) api.Incident {
	out := api.Incident{
		Id:            inc.ID,
		Number:        inc.Number,
		Name:          inc.Name,
		Summary:       inc.Summary,
		SeverityLevel: inc.SeverityLevel,
		Status:        api.IncidentStatus(inc.Status),
		DeclaredAt:    inc.DeclaredAt,
		CustomFields:  inc.CustomFields,
		Services:      servicesDTO(inc.Services),
		Alerts:        alertRefsDTO(inc.Alerts),
		Relations:     relationRefsDTO(inc.Related),
		Followups:     followupsDTO(inc.Followups),
		Timeline:      eventsDTO(inc.Timeline),
	}
	if out.CustomFields == nil {
		out.CustomFields = map[string]string{}
	}
	if inc.LeadUserID != "" {
		out.LeadUserId = ptr(inc.LeadUserID)
	}
	if inc.LeadLabel != "" {
		out.LeadLabel = ptr(inc.LeadLabel)
	}
	if inc.TeamSlug != "" {
		out.TeamSlug = ptr(inc.TeamSlug)
	}
	if inc.ResolutionSummary != "" {
		out.ResolutionSummary = ptr(inc.ResolutionSummary)
	}
	if inc.DeclaredBy != "" {
		out.DeclaredBy = ptr(inc.DeclaredBy)
	}
	if !inc.ResolvedAt.IsZero() {
		out.ResolvedAt = ptr(inc.ResolvedAt)
	}
	return out
}

func serviceDTO(s entity.Service) api.Service {
	out := api.Service{Id: s.ID, Slug: s.Slug, Name: s.Name, Description: s.Description}
	if s.TeamSlug != "" {
		out.TeamSlug = ptr(s.TeamSlug)
	}
	return out
}

func servicesDTO(in []entity.Service) []api.Service {
	out := make([]api.Service, 0, len(in))
	for _, s := range in {
		out = append(out, serviceDTO(s))
	}
	return out
}

func alertRefsDTO(in []entity.IncidentAlert) []api.IncidentAlertRef {
	out := make([]api.IncidentAlertRef, 0, len(in))
	for _, a := range in {
		out = append(out, api.IncidentAlertRef{AlertId: a.AlertID, Title: a.Title, Severity: string(a.Severity), Status: string(a.Status)})
	}
	return out
}

func relationRefsDTO(in []entity.IncidentRelation) []api.IncidentRelationRef {
	out := make([]api.IncidentRelationRef, 0, len(in))
	for _, r := range in {
		out = append(out, api.IncidentRelationRef{
			Id:            r.ID,
			Kind:          api.IncidentRelationRefKind(r.Kind),
			RelatedId:     r.RelatedID,
			RelatedNumber: r.RelatedNumber,
			RelatedName:   r.RelatedName,
			RelatedStatus: string(r.RelatedStatus),
		})
	}
	return out
}

func followupDTO(f entity.IncidentFollowup) api.IncidentFollowup {
	out := api.IncidentFollowup{Id: f.ID, IncidentId: f.IncidentID, Title: f.Title, Done: f.Done}
	if f.OwnerUserID != "" {
		out.OwnerUserId = ptr(f.OwnerUserID)
	}
	if f.OwnerLabel != "" {
		out.OwnerLabel = ptr(f.OwnerLabel)
	}
	if !f.DueAt.IsZero() {
		out.DueAt = ptr(f.DueAt)
	}
	if !f.DoneAt.IsZero() {
		out.DoneAt = ptr(f.DoneAt)
	}
	return out
}

func followupsDTO(in []entity.IncidentFollowup) []api.IncidentFollowup {
	out := make([]api.IncidentFollowup, 0, len(in))
	for _, f := range in {
		out = append(out, followupDTO(f))
	}
	return out
}

func eventsDTO(in []entity.IncidentEvent) []api.IncidentEvent {
	out := make([]api.IncidentEvent, 0, len(in))
	for _, e := range in {
		ev := api.IncidentEvent{Id: e.ID, At: e.At, Kind: e.Kind, Text: e.Text}
		if e.Actor != "" {
			ev.Actor = ptr(e.Actor)
		}
		out = append(out, ev)
	}
	return out
}

func severityDTO(s entity.IncidentSeverity) api.IncidentSeverity {
	out := api.IncidentSeverity{Id: ptr(s.ID), Level: s.Level, Label: s.Label, Tone: s.Tone, Position: ptr(s.Position)}
	if s.Definition != "" {
		out.Definition = ptr(s.Definition)
	}
	return out
}

func fieldDTO(f entity.IncidentFieldDef) api.IncidentField {
	options := append([]string{}, f.Options...)
	return api.IncidentField{
		Id:       ptr(f.ID),
		Slug:     ptr(f.Slug),
		Name:     f.Name,
		Kind:     api.IncidentFieldKind(f.Kind),
		Options:  &options,
		Position: ptr(f.Position),
	}
}

func (h *handler) ListIncidents(ctx context.Context, request api.ListIncidentsRequestObject) (api.ListIncidentsResponseObject, error) {
	filter := entity.IncidentFilter{}
	if request.Params.Status != nil {
		for _, s := range *request.Params.Status {
			filter.Statuses = append(filter.Statuses, entity.IncidentStatus(s))
		}
	}
	if request.Params.Severity != nil {
		filter.Severities = *request.Params.Severity
	}
	if request.Params.Service != nil {
		filter.ServiceIDs = *request.Params.Service
	}
	if request.Params.Team != nil {
		filter.TeamIDs = *request.Params.Team
	}
	if request.Params.Active != nil {
		filter.ActiveOnly = *request.Params.Active
	}
	if request.Params.Since != nil {
		filter.Since = *request.Params.Since
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
	page, err := h.incidents.List(ctx, request.WorkspaceId, filter)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListIncidents401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListIncidents403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListIncidents404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.Incident, 0, len(page.Incidents))
	for _, inc := range page.Incidents {
		items = append(items, h.incidentDTO(inc))
	}
	out := api.ListIncidents200JSONResponse{Items: items}
	if page.NextCursor != "" {
		out.NextCursor = ptr(page.NextCursor)
	}
	return out, nil
}

func (h *handler) GetIncident(ctx context.Context, request api.GetIncidentRequestObject) (api.GetIncidentResponseObject, error) {
	inc, err := h.incidents.Get(ctx, request.WorkspaceId, request.IncidentId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.GetIncident401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.GetIncident403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.GetIncident404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.GetIncident200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) DeclareIncident(ctx context.Context, request api.DeclareIncidentRequestObject) (api.DeclareIncidentResponseObject, error) {
	in := entity.IncidentDeclare{}
	if request.Body != nil {
		in.Name = request.Body.Name
		in.Summary = derefString(request.Body.Summary)
		in.SeverityLevel = derefString(request.Body.SeverityLevel)
		in.TeamSlug = derefString(request.Body.TeamSlug)
		in.LeadUserID = derefString(request.Body.LeadUserId)
		in.ServiceIDs = derefStrings(request.Body.ServiceIds)
	}
	inc, err := h.incidents.Declare(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.DeclareIncident400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.DeclareIncident401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeclareIncident403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeclareIncident404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeclareIncident201JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) DeclareIncidentFromAlert(ctx context.Context, request api.DeclareIncidentFromAlertRequestObject) (api.DeclareIncidentFromAlertResponseObject, error) {
	in := entity.IncidentDeclare{}
	if request.Body != nil {
		in.FromAlertID = request.Body.AlertId
		in.Name = derefString(request.Body.Name)
		in.Summary = derefString(request.Body.Summary)
		in.SeverityLevel = derefString(request.Body.SeverityLevel)
		in.TeamSlug = derefString(request.Body.TeamSlug)
		in.LeadUserID = derefString(request.Body.LeadUserId)
		in.ServiceIDs = derefStrings(request.Body.ServiceIds)
	}
	inc, err := h.incidents.Declare(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.DeclareIncidentFromAlert400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.DeclareIncidentFromAlert401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeclareIncidentFromAlert403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeclareIncidentFromAlert404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeclareIncidentFromAlert201JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) UpdateIncident(ctx context.Context, request api.UpdateIncidentRequestObject) (api.UpdateIncidentResponseObject, error) {
	in := entity.IncidentUpdate{}
	if request.Body != nil {
		in.Name = request.Body.Name
		in.Summary = derefString(request.Body.Summary)
		in.TeamSlug = derefString(request.Body.TeamSlug)
		in.LeadUserID = derefString(request.Body.LeadUserId)
		in.ServiceIDs = derefStrings(request.Body.ServiceIds)
	}
	inc, err := h.incidents.Update(ctx, request.WorkspaceId, request.IncidentId, in)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateIncident400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateIncident401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateIncident403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateIncident404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateIncident200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) ChangeIncidentStatus(ctx context.Context, request api.ChangeIncidentStatusRequestObject) (api.ChangeIncidentStatusResponseObject, error) {
	to := entity.IncidentStatus("")
	if request.Body != nil {
		to = entity.IncidentStatus(request.Body.Status)
	}
	inc, err := h.incidents.ChangeStatus(ctx, request.WorkspaceId, request.IncidentId, to)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.ChangeIncidentStatus400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.ChangeIncidentStatus401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ChangeIncidentStatus403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ChangeIncidentStatus404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.ChangeIncidentStatus409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ChangeIncidentStatus200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) ChangeIncidentSeverity(ctx context.Context, request api.ChangeIncidentSeverityRequestObject) (api.ChangeIncidentSeverityResponseObject, error) {
	level := ""
	if request.Body != nil {
		level = request.Body.Level
	}
	inc, err := h.incidents.ChangeSeverity(ctx, request.WorkspaceId, request.IncidentId, level)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.ChangeIncidentSeverity400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.ChangeIncidentSeverity401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ChangeIncidentSeverity403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ChangeIncidentSeverity404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ChangeIncidentSeverity200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) ResolveIncident(ctx context.Context, request api.ResolveIncidentRequestObject) (api.ResolveIncidentResponseObject, error) {
	summary := ""
	if request.Body != nil {
		summary = request.Body.Summary
	}
	inc, err := h.incidents.Resolve(ctx, request.WorkspaceId, request.IncidentId, summary)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.ResolveIncident400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.ResolveIncident401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ResolveIncident403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ResolveIncident404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.ResolveIncident409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ResolveIncident200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) ReopenIncident(ctx context.Context, request api.ReopenIncidentRequestObject) (api.ReopenIncidentResponseObject, error) {
	inc, err := h.incidents.Reopen(ctx, request.WorkspaceId, request.IncidentId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ReopenIncident401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ReopenIncident403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ReopenIncident404ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusConflict:
			return api.ReopenIncident409ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ReopenIncident200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) SetIncidentCustomFields(ctx context.Context, request api.SetIncidentCustomFieldsRequestObject) (api.SetIncidentCustomFieldsResponseObject, error) {
	fields := map[string]string{}
	if request.Body != nil {
		fields = request.Body.Fields
	}
	inc, err := h.incidents.SetCustomFields(ctx, request.WorkspaceId, request.IncidentId, fields)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.SetIncidentCustomFields400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.SetIncidentCustomFields401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.SetIncidentCustomFields403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.SetIncidentCustomFields404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.SetIncidentCustomFields200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) LinkIncidentAlert(ctx context.Context, request api.LinkIncidentAlertRequestObject) (api.LinkIncidentAlertResponseObject, error) {
	alertID := ""
	if request.Body != nil {
		alertID = request.Body.AlertId
	}
	inc, err := h.incidents.LinkAlert(ctx, request.WorkspaceId, request.IncidentId, alertID)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.LinkIncidentAlert400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.LinkIncidentAlert401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.LinkIncidentAlert403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.LinkIncidentAlert404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.LinkIncidentAlert200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) UnlinkIncidentAlert(ctx context.Context, request api.UnlinkIncidentAlertRequestObject) (api.UnlinkIncidentAlertResponseObject, error) {
	inc, err := h.incidents.UnlinkAlert(ctx, request.WorkspaceId, request.IncidentId, request.AlertId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.UnlinkIncidentAlert401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UnlinkIncidentAlert403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UnlinkIncidentAlert404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UnlinkIncidentAlert200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) RelateIncident(ctx context.Context, request api.RelateIncidentRequestObject) (api.RelateIncidentResponseObject, error) {
	relatedID := ""
	kind := entity.IncidentRelationKind("")
	if request.Body != nil {
		relatedID = request.Body.RelatedId
		kind = entity.IncidentRelationKind(request.Body.Kind)
	}
	inc, err := h.incidents.Relate(ctx, request.WorkspaceId, request.IncidentId, relatedID, kind)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.RelateIncident400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.RelateIncident401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.RelateIncident403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.RelateIncident404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.RelateIncident200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) UnrelateIncident(ctx context.Context, request api.UnrelateIncidentRequestObject) (api.UnrelateIncidentResponseObject, error) {
	inc, err := h.incidents.Unrelate(ctx, request.WorkspaceId, request.IncidentId, request.RelationId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.UnrelateIncident401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UnrelateIncident403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UnrelateIncident404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UnrelateIncident200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) AddIncidentFollowup(ctx context.Context, request api.AddIncidentFollowupRequestObject) (api.AddIncidentFollowupResponseObject, error) {
	in := entity.NewFollowup{}
	if request.Body != nil {
		in.Title = request.Body.Title
		in.OwnerUserID = derefString(request.Body.OwnerUserId)
		if request.Body.DueAt != nil {
			in.DueAt = *request.Body.DueAt
		}
	}
	inc, err := h.incidents.AddFollowup(ctx, request.WorkspaceId, request.IncidentId, in)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.AddIncidentFollowup400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.AddIncidentFollowup401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.AddIncidentFollowup403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.AddIncidentFollowup404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.AddIncidentFollowup200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) ToggleIncidentFollowup(ctx context.Context, request api.ToggleIncidentFollowupRequestObject) (api.ToggleIncidentFollowupResponseObject, error) {
	done := false
	if request.Body != nil {
		done = request.Body.Done
	}
	inc, err := h.incidents.ToggleFollowup(ctx, request.WorkspaceId, request.IncidentId, request.FollowupId, done)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.ToggleIncidentFollowup400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.ToggleIncidentFollowup401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ToggleIncidentFollowup403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ToggleIncidentFollowup404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ToggleIncidentFollowup200JSONResponse(h.incidentDTO(inc)), nil
}

func (h *handler) ListIncidentFollowups(ctx context.Context, request api.ListIncidentFollowupsRequestObject) (api.ListIncidentFollowupsResponseObject, error) {
	list, err := h.incidents.ListOpenFollowups(ctx, request.WorkspaceId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListIncidentFollowups401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListIncidentFollowups403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListIncidentFollowups404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	items := make([]api.IncidentFollowup, 0, len(list))
	for _, f := range list {
		items = append(items, followupDTO(f))
	}
	return api.ListIncidentFollowups200JSONResponse{Items: items}, nil
}

func (h *handler) ListIncidentSeverities(ctx context.Context, request api.ListIncidentSeveritiesRequestObject) (api.ListIncidentSeveritiesResponseObject, error) {
	list, err := h.incidents.ListSeverities(ctx, request.WorkspaceId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListIncidentSeverities401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListIncidentSeverities403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListIncidentSeverities404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ListIncidentSeverities200JSONResponse{Items: severitiesDTO(list)}, nil
}

func (h *handler) SaveIncidentSeverities(ctx context.Context, request api.SaveIncidentSeveritiesRequestObject) (api.SaveIncidentSeveritiesResponseObject, error) {
	in := make([]entity.IncidentSeverity, 0)
	if request.Body != nil {
		for _, s := range request.Body.Severities {
			in = append(in, entity.IncidentSeverity{
				Level:      s.Level,
				Label:      s.Label,
				Definition: derefString(s.Definition),
				Tone:       s.Tone,
			})
		}
	}
	list, err := h.incidents.SaveSeverities(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.SaveIncidentSeverities400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.SaveIncidentSeverities401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.SaveIncidentSeverities403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.SaveIncidentSeverities404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.SaveIncidentSeverities200JSONResponse{Items: severitiesDTO(list)}, nil
}

func (h *handler) ListIncidentFields(ctx context.Context, request api.ListIncidentFieldsRequestObject) (api.ListIncidentFieldsResponseObject, error) {
	list, err := h.incidents.ListFieldDefs(ctx, request.WorkspaceId)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListIncidentFields401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListIncidentFields403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListIncidentFields404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ListIncidentFields200JSONResponse{Items: fieldsDTO(list)}, nil
}

func (h *handler) SaveIncidentFields(ctx context.Context, request api.SaveIncidentFieldsRequestObject) (api.SaveIncidentFieldsResponseObject, error) {
	in := make([]entity.IncidentFieldDef, 0)
	if request.Body != nil {
		for _, f := range request.Body.Fields {
			def := entity.IncidentFieldDef{
				ID:   derefString(f.Id),
				Slug: derefString(f.Slug),
				Name: f.Name,
				Kind: entity.CustomFieldKind(f.Kind),
			}
			if f.Options != nil {
				def.Options = *f.Options
			}
			in = append(in, def)
		}
	}
	list, err := h.incidents.SaveFieldDefs(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := incidentProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.SaveIncidentFields400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.SaveIncidentFields401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.SaveIncidentFields403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.SaveIncidentFields404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.SaveIncidentFields200JSONResponse{Items: fieldsDTO(list)}, nil
}

func severitiesDTO(in []entity.IncidentSeverity) []api.IncidentSeverity {
	out := make([]api.IncidentSeverity, 0, len(in))
	for _, s := range in {
		out = append(out, severityDTO(s))
	}
	return out
}

func fieldsDTO(in []entity.IncidentFieldDef) []api.IncidentField {
	out := make([]api.IncidentField, 0, len(in))
	for _, f := range in {
		out = append(out, fieldDTO(f))
	}
	return out
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefStrings(v *[]string) []string {
	if v == nil {
		return nil
	}
	return *v
}
