package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) GetMe(ctx context.Context, _ api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	u, err := h.auth.Profile(ctx)
	if err != nil {
		return nil, err
	}
	return api.GetMe200JSONResponse{
		Id:               u.ID,
		Name:             u.Name,
		Email:            u.Email,
		Timezone:         u.Timezone,
		TwoFactorEnabled: u.TOTPEnabled,
	}, nil
}

func (h *handler) UpdateMe(ctx context.Context, request api.UpdateMeRequestObject) (api.UpdateMeResponseObject, error) {
	if request.Body == nil {
		return api.UpdateMe400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	u, err := h.users.UpdateProfile(ctx, entity.ProfileUpdate{Name: request.Body.Name, Timezone: request.Body.Timezone})
	if err != nil {
		if isValidation(err) {
			return api.UpdateMe400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid profile", validationDetail(err), "")), nil
		}
		return nil, err
	}
	return api.UpdateMe200JSONResponse{Id: u.ID, Name: u.Name, Email: u.Email, Timezone: u.Timezone, TwoFactorEnabled: u.TOTPEnabled}, nil
}

func (h *handler) ChangePassword(ctx context.Context, request api.ChangePasswordRequestObject) (api.ChangePasswordResponseObject, error) {
	if request.Body == nil {
		return api.ChangePassword400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	err := h.users.ChangePassword(ctx, request.Body.CurrentPassword, request.Body.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCredentials):
			return api.ChangePassword400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Wrong current password", "The current password is incorrect. Check it and try again.", "")), nil
		case errors.Is(err, entity.ErrUserWeakPassword):
			return api.ChangePassword400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Weak password", validationDetail(err), "")), nil
		default:
			return nil, err
		}
	}
	return api.ChangePassword204Response{}, nil
}

func (h *handler) ListSessions(ctx context.Context, _ api.ListSessionsRequestObject) (api.ListSessionsResponseObject, error) {
	sessions, err := h.users.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	id, _ := entity.IdentityFrom(ctx)
	items := make([]api.Session, 0, len(sessions))
	for _, s := range sessions {
		item := api.Session{Id: s.ID, CreatedAt: s.CreatedAt, LastSeenAt: s.LastSeenAt, Current: s.ID == id.SessionID}
		if s.IP != "" {
			item.Ip = ptr(s.IP)
		}
		if s.UserAgent != "" {
			item.UserAgent = ptr(s.UserAgent)
		}
		items = append(items, item)
	}
	return api.ListSessions200JSONResponse{Items: items}, nil
}

func (h *handler) RevokeSession(ctx context.Context, request api.RevokeSessionRequestObject) (api.RevokeSessionResponseObject, error) {
	err := h.users.RevokeSession(ctx, request.SessionId)
	if err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			return api.RevokeSession404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such session.", "")), nil
		}
		return nil, err
	}
	return api.RevokeSession204Response{}, nil
}
