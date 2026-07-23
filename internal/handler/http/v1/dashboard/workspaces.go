package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) ListWorkspaces(ctx context.Context, _ api.ListWorkspacesRequestObject) (api.ListWorkspacesResponseObject, error) {
	list, err := h.workspaces.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]api.Workspace, 0, len(list))
	for _, w := range list {
		items = append(items, workspaceDTO(w))
	}
	return api.ListWorkspaces200JSONResponse{Items: items}, nil
}

func (h *handler) GetWorkspace(ctx context.Context, request api.GetWorkspaceRequestObject) (api.GetWorkspaceResponseObject, error) {
	ws, err := h.workspaces.Get(ctx, request.WorkspaceId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.GetWorkspace404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Workspace not found",
				"No workspace with that identifier, or you're not a member of it.", "")), nil
		default:
			return nil, err
		}
	}
	return api.GetWorkspace200JSONResponse(workspaceDTO(ws)), nil
}

func workspaceDTO(w entity.Workspace) api.Workspace {
	dto := api.Workspace{Id: w.Slug, Name: w.Name, Timezone: w.Timezone, Environment: ptr(w.Environment)}
	if w.Role != "" {
		role := api.Role(w.Role)
		dto.Role = &role
	}
	return dto
}
