package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) ListAudit(ctx context.Context, request api.ListAuditRequestObject) (api.ListAuditResponseObject, error) {
	filter := entity.AuditFilter{
		Cursor:       strParam(request.Params.Cursor),
		Limit:        intParam(request.Params.Limit),
		ActorUserID:  strParam(request.Params.Actor),
		ActionPrefix: strParam(request.Params.Action),
		Query:        strParam(request.Params.Q),
	}
	page, err := h.audits.List(ctx, request.WorkspaceId, filter)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ListAudit403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "You don't have permission to read the audit trail.", "")), nil
		case errors.Is(err, entity.ErrAuditInvalidCursor):
			return api.ListAudit400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid cursor", "That pagination cursor isn't valid.", "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ListAudit404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		default:
			return nil, err
		}
	}
	items := make([]api.AuditEvent, 0, len(page.Events))
	for _, e := range page.Events {
		items = append(items, auditDTO(e))
	}
	resp := api.ListAudit200JSONResponse{Items: items}
	if page.NextCursor != "" {
		resp.NextCursor = ptr(page.NextCursor)
	}
	return resp, nil
}

func auditDTO(e entity.AuditEvent) api.AuditEvent {
	return api.AuditEvent{
		Id:     e.ID,
		At:     e.At,
		Actor:  e.ActorLabel,
		Action: e.Action,
		Target: e.Target,
		Ip:     e.IP,
	}
}

func strParam(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func intParam(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
