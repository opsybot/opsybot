package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func serviceProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to services in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case isValidation(err):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Invalid service", validationDetail(err), "")
	case errors.Is(err, entity.ErrServiceSlugTaken):
		return http.StatusConflict, prob(http.StatusConflict, "Name taken", "A service already goes by that name.", "")
	case errors.Is(err, entity.ErrTeamNotFound):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Unknown team", "Pick an existing team for this service.", "")
	case errors.Is(err, entity.ErrServiceNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That service does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func (h *handler) ListServices(ctx context.Context, request api.ListServicesRequestObject) (api.ListServicesResponseObject, error) {
	list, err := h.services.List(ctx, request.WorkspaceId)
	if err != nil {
		status, p := serviceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListServices401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListServices403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListServices404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.ListServices200JSONResponse{Items: servicesDTO(list)}, nil
}

func (h *handler) CreateService(ctx context.Context, request api.CreateServiceRequestObject) (api.CreateServiceResponseObject, error) {
	in := entity.NewService{}
	if request.Body != nil {
		in.Name = request.Body.Name
		in.TeamSlug = derefString(request.Body.TeamSlug)
		in.Description = derefString(request.Body.Description)
	}
	svc, err := h.services.Create(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := serviceProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.CreateService400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.CreateService401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.CreateService403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.CreateService404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.CreateService201JSONResponse(serviceDTO(svc)), nil
}

func (h *handler) UpdateService(ctx context.Context, request api.UpdateServiceRequestObject) (api.UpdateServiceResponseObject, error) {
	in := entity.ServiceUpdate{}
	if request.Body != nil {
		in.Name = request.Body.Name
		in.TeamSlug = derefString(request.Body.TeamSlug)
		in.Description = derefString(request.Body.Description)
	}
	svc, err := h.services.Update(ctx, request.WorkspaceId, request.ServiceId, in)
	if err != nil {
		status, p := serviceProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.UpdateService400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.UpdateService401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.UpdateService403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.UpdateService404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.UpdateService200JSONResponse(serviceDTO(svc)), nil
}

func (h *handler) DeleteService(ctx context.Context, request api.DeleteServiceRequestObject) (api.DeleteServiceResponseObject, error) {
	err := h.services.Delete(ctx, request.WorkspaceId, request.ServiceId)
	if err != nil {
		status, p := serviceProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DeleteService401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeleteService403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeleteService404ApplicationProblemPlusJSONResponse(p), nil
		}
		return nil, err
	}
	return api.DeleteService204Response{}, nil
}
