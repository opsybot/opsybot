package dashboard

import (
	"context"
	"errors"
	"net/http"
	"sort"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func routeProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to alert routing in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid rule", validationDetail(err), "")
	case errors.Is(err, entity.ErrAlertRouteNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That routing rule does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func routeDTO(r entity.AlertRoute) api.AlertRoute {
	conds := make([]api.RouteCondition, 0, len(r.Conditions))
	for _, c := range r.Conditions {
		conds = append(conds, api.RouteCondition{Field: c.Field, Op: api.RouteConditionOp(c.Op), Value: c.Value})
	}
	return api.AlertRoute{Id: r.ID, Position: r.Position, PolicyRef: r.PolicyRef, Conditions: conds}
}

func routeInput(policyRef string, conditions []api.RouteCondition) entity.NewAlertRoute {
	in := entity.NewAlertRoute{PolicyRef: policyRef}
	for _, c := range conditions {
		in.Conditions = append(in.Conditions, entity.RouteCondition{
			Field: c.Field, Op: entity.ConditionOp(c.Op), Value: c.Value,
		})
	}
	return in
}

func (h *handler) ListAlertRoutes(ctx context.Context, request api.ListAlertRoutesRequestObject) (api.ListAlertRoutesResponseObject, error) {
	routes, settings, err := h.routes.List(ctx, request.WorkspaceId)
	if err != nil {
		status, p := routeProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListAlertRoutes401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListAlertRoutes403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListAlertRoutes404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}

	items := make([]api.AlertRoute, 0, len(routes))
	known := map[string]struct{}{settings.DefaultPolicyRef: {}}
	for _, r := range routes {
		items = append(items, routeDTO(r))
		known[r.PolicyRef] = struct{}{}
	}
	refs := make([]string, 0, len(known))
	for ref := range known {
		if ref != "" {
			refs = append(refs, ref)
		}
	}
	sort.Strings(refs)

	return api.ListAlertRoutes200JSONResponse{
		Items:            items,
		DefaultPolicyRef: settings.DefaultPolicyRef,
		KnownPolicyRefs:  refs,
	}, nil
}

func (h *handler) CreateAlertRoute(ctx context.Context, request api.CreateAlertRouteRequestObject) (api.CreateAlertRouteResponseObject, error) {
	created, err := h.routes.Create(ctx, request.WorkspaceId, routeInput(request.Body.PolicyRef, request.Body.Conditions))
	if err != nil {
		status, p := routeProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.CreateAlertRoute400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.CreateAlertRoute401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.CreateAlertRoute403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.CreateAlertRoute404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.CreateAlertRoute201JSONResponse(routeDTO(created)), nil
}

func (h *handler) UpdateAlertRoute(ctx context.Context, request api.UpdateAlertRouteRequestObject) (api.UpdateAlertRouteResponseObject, error) {
	updated, err := h.routes.Update(ctx, request.WorkspaceId, request.RouteId, routeInput(request.Body.PolicyRef, request.Body.Conditions))
	if err != nil {
		status, p := routeProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateAlertRoute400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateAlertRoute401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateAlertRoute403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateAlertRoute404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateAlertRoute200JSONResponse(routeDTO(updated)), nil
}

func (h *handler) DeleteAlertRoute(ctx context.Context, request api.DeleteAlertRouteRequestObject) (api.DeleteAlertRouteResponseObject, error) {
	if err := h.routes.Delete(ctx, request.WorkspaceId, request.RouteId); err != nil {
		status, p := routeProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DeleteAlertRoute401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeleteAlertRoute403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeleteAlertRoute404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeleteAlertRoute204Response{}, nil
}

func (h *handler) ReorderAlertRoutes(ctx context.Context, request api.ReorderAlertRoutesRequestObject) (api.ReorderAlertRoutesResponseObject, error) {
	if err := h.routes.Reorder(ctx, request.WorkspaceId, request.Body.Ids); err != nil {
		status, p := routeProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.ReorderAlertRoutes400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.ReorderAlertRoutes401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ReorderAlertRoutes403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ReorderAlertRoutes404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ReorderAlertRoutes204Response{}, nil
}

func (h *handler) SetDefaultAlertPolicy(ctx context.Context, request api.SetDefaultAlertPolicyRequestObject) (api.SetDefaultAlertPolicyResponseObject, error) {
	if err := h.routes.SetDefaultPolicy(ctx, request.WorkspaceId, request.Body.PolicyRef); err != nil {
		status, p := routeProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.SetDefaultAlertPolicy400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.SetDefaultAlertPolicy401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.SetDefaultAlertPolicy403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.SetDefaultAlertPolicy404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.SetDefaultAlertPolicy204Response{}, nil
}
