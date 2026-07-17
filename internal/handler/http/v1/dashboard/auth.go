package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) GetSetupStatus(ctx context.Context, _ api.GetSetupStatusRequestObject) (api.GetSetupStatusResponseObject, error) {
	required, err := h.auth.SetupRequired(ctx)
	if err != nil {
		return nil, err
	}
	return api.GetSetupStatus200JSONResponse{Required: required}, nil
}

func (h *handler) Setup(ctx context.Context, request api.SetupRequestObject) (api.SetupResponseObject, error) {
	if request.Body == nil {
		return api.Setup400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	b := request.Body
	info := entity.RequestInfoFrom(ctx)
	res, err := h.auth.Setup(ctx, entity.Setup{
		UserName:      b.Name,
		Email:         b.Email,
		Password:      b.Password,
		WorkspaceName: b.Workspace,
		WorkspaceSlug: entity.Slugify(b.Workspace),
		Timezone:      b.Timezone,
	}, info.IP, info.UserAgent)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrSetupAlreadyDone):
			return api.Setup409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Already set up",
				"This instance is already set up. Setup runs only once. Sign in at /login instead.", "already-set-up")), nil
		case isValidation(err):
			return api.Setup400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid setup details",
				validationDetail(err), "")), nil
		default:
			return nil, err
		}
	}
	return api.Setup201JSONResponse{
		Body:    sessionUser(res.User),
		Headers: api.Setup201ResponseHeaders{SetCookie: ptr(h.sessionCookie(res.Token, res.Session.ExpiresAt))},
	}, nil
}

func (h *handler) Login(ctx context.Context, request api.LoginRequestObject) (api.LoginResponseObject, error) {
	if request.Body == nil {
		return api.Login400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	b := request.Body
	info := entity.RequestInfoFrom(ctx)
	remember := b.Remember != nil && *b.Remember
	res, err := h.auth.Login(ctx, entity.LoginInput{
		Email: b.Email, Password: b.Password, IP: info.IP, UserAgent: info.UserAgent, Remember: remember,
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCredentials):
			return api.Login401ApplicationProblemPlusJSONResponse(prob(http.StatusUnauthorized, "Sign-in failed",
				"The email or password is incorrect. Check both and try again, or use 'Forgot password'.", "")), nil
		case errors.Is(err, entity.ErrUserDeactivated):
			return api.Login403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Account deactivated",
				"Your account is deactivated. Contact your workspace admin to be reactivated.", "deactivated")), nil
		case errors.Is(err, entity.ErrSSORequired):
			return api.Login403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Single sign-on required",
				"Password sign-in is turned off for this account. Continue with your identity provider instead.", "sso-required")), nil
		default:
			return nil, err
		}
	}
	return api.Login200JSONResponse{
		Body:    api.LoginResult{Status: api.LoginResultStatusOk, User: ptr(sessionUser(res.User))},
		Headers: api.Login200ResponseHeaders{SetCookie: ptr(h.sessionCookie(res.Token, res.Session.ExpiresAt))},
	}, nil
}

func (h *handler) Logout(ctx context.Context, _ api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	if err := h.auth.Logout(ctx); err != nil {
		if errors.Is(err, entity.ErrUnauthenticated) {
			return api.Logout401ApplicationProblemPlusJSONResponse(prob(http.StatusUnauthorized, "Not authenticated",
				"You're not signed in.", "")), nil
		}
		return nil, err
	}
	return api.Logout204Response{Headers: api.Logout204ResponseHeaders{SetCookie: ptr(h.clearCookie())}}, nil
}

func sessionUser(u entity.User) api.SessionUser {
	return api.SessionUser{Id: u.ID, Name: u.Name, Email: u.Email, Timezone: u.Timezone}
}
