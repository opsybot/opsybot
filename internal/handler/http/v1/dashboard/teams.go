package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) ListTeams(ctx context.Context, request api.ListTeamsRequestObject) (api.ListTeamsResponseObject, error) {
	includeArchived := request.Params.IncludeArchived != nil && *request.Params.IncludeArchived
	list, err := h.teams.List(ctx, request.WorkspaceId, includeArchived)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ListTeams403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "You don't have permission to view teams.", "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ListTeams404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		default:
			return nil, err
		}
	}
	items := make([]api.Team, 0, len(list))
	for _, t := range list {
		items = append(items, teamDTO(t))
	}
	return api.ListTeams200JSONResponse{Items: items}, nil
}

func (h *handler) CreateTeam(ctx context.Context, request api.CreateTeamRequestObject) (api.CreateTeamResponseObject, error) {
	if request.Body == nil {
		return api.CreateTeam400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	team, err := h.teams.Create(ctx, request.WorkspaceId, entity.NewTeam{Name: request.Body.Name, MemberIDs: memberIDs(request.Body.MemberIds)})
	if err != nil {
		if resp := createTeamError(err); resp != nil {
			return resp, nil
		}
		return nil, err
	}
	return api.CreateTeam201JSONResponse(teamDTO(team)), nil
}

func (h *handler) GetTeam(ctx context.Context, request api.GetTeamRequestObject) (api.GetTeamResponseObject, error) {
	team, err := h.teams.Get(ctx, request.WorkspaceId, request.TeamSlug)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.GetTeam403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "You don't have permission to view teams.", "")), nil
		case errors.Is(err, entity.ErrTeamNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.GetTeam404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such team.", "")), nil
		default:
			return nil, err
		}
	}
	return api.GetTeam200JSONResponse(teamDTO(team)), nil
}

func (h *handler) UpdateTeam(ctx context.Context, request api.UpdateTeamRequestObject) (api.UpdateTeamResponseObject, error) {
	if request.Body == nil {
		return api.UpdateTeam400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	team, err := h.teams.Update(ctx, request.WorkspaceId, request.TeamSlug, entity.TeamUpdate{Name: request.Body.Name, MemberIDs: memberIDs(request.Body.MemberIds)})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.UpdateTeam403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can change teams.", "")), nil
		case errors.Is(err, entity.ErrTeamNameInvalid), errors.Is(err, entity.ErrTeamTooManyMembers), errors.Is(err, entity.ErrTeamMemberInvalid):
			return api.UpdateTeam400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid team", teamValidationDetail(err), "")), nil
		case errors.Is(err, entity.ErrTeamNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.UpdateTeam404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such team.", "")), nil
		case errors.Is(err, entity.ErrTeamArchived):
			return api.UpdateTeam409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Team archived", "Restore this team before editing it.", "")), nil
		default:
			return nil, err
		}
	}
	return api.UpdateTeam200JSONResponse(teamDTO(team)), nil
}

func (h *handler) ArchiveTeam(ctx context.Context, request api.ArchiveTeamRequestObject) (api.ArchiveTeamResponseObject, error) {
	team, err := h.teams.Archive(ctx, request.WorkspaceId, request.TeamSlug)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ArchiveTeam403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can archive teams.", "")), nil
		case errors.Is(err, entity.ErrTeamNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ArchiveTeam404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such team.", "")), nil
		case errors.Is(err, entity.ErrTeamArchived):
			return api.ArchiveTeam409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Already archived", "That team is already archived.", "")), nil
		default:
			return nil, err
		}
	}
	return api.ArchiveTeam200JSONResponse(teamDTO(team)), nil
}

func (h *handler) UnarchiveTeam(ctx context.Context, request api.UnarchiveTeamRequestObject) (api.UnarchiveTeamResponseObject, error) {
	team, err := h.teams.Unarchive(ctx, request.WorkspaceId, request.TeamSlug)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.UnarchiveTeam403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can restore teams.", "")), nil
		case errors.Is(err, entity.ErrTeamNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.UnarchiveTeam404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such team.", "")), nil
		case errors.Is(err, entity.ErrTeamNotArchived):
			return api.UnarchiveTeam409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Not archived", "That team is not archived.", "")), nil
		default:
			return nil, err
		}
	}
	return api.UnarchiveTeam200JSONResponse(teamDTO(team)), nil
}

func createTeamError(err error) api.CreateTeamResponseObject {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return api.CreateTeam403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can create teams.", ""))
	case errors.Is(err, entity.ErrTeamNameInvalid), errors.Is(err, entity.ErrTeamTooManyMembers), errors.Is(err, entity.ErrTeamMemberInvalid):
		return api.CreateTeam400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid team", teamValidationDetail(err), ""))
	case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return api.CreateTeam404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", ""))
	case errors.Is(err, entity.ErrTeamSlugTaken):
		return api.CreateTeam409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Name taken", "A team with a similar name already exists. Pick a different name.", ""))
	default:
		return nil
	}
}

func teamValidationDetail(err error) string {
	switch {
	case errors.Is(err, entity.ErrTeamTooManyMembers):
		return "A team can have at most 50 members."
	case errors.Is(err, entity.ErrTeamMemberInvalid):
		return "Every team member must be an active member of this workspace."
	default:
		return "Give the team a name of 60 characters or fewer."
	}
}

func memberIDs(ids *[]string) []string {
	if ids == nil {
		return nil
	}
	return *ids
}

func teamDTO(t entity.Team) api.Team {
	return api.Team{
		Id:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		MemberIds: t.MemberIDs,
		Archived:  t.Archived,
		CreatedAt: t.CreatedAt,
	}
}
