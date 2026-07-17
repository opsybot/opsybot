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
	if res.Outcome == entity.LoginOutcomeTwoFactor {
		return api.Login200JSONResponse{
			Body:    api.LoginResult{Status: api.LoginResultStatusTwoFactorRequired},
			Headers: api.Login200ResponseHeaders{SetCookie: ptr(h.pendingCookie(res.PendingToken))},
		}, nil
	}
	return api.Login200JSONResponse{
		Body:    api.LoginResult{Status: api.LoginResultStatusOk, User: ptr(sessionUser(res.User))},
		Headers: api.Login200ResponseHeaders{SetCookie: ptr(h.sessionCookie(res.Token, res.Session.ExpiresAt))},
	}, nil
}

func (h *handler) VerifyTwoFactor(ctx context.Context, request api.VerifyTwoFactorRequestObject) (api.VerifyTwoFactorResponseObject, error) {
	pending := pendingFrom(ctx)
	if request.Body == nil || pending == "" {
		return api.VerifyTwoFactor401ApplicationProblemPlusJSONResponse(prob(http.StatusUnauthorized, "Sign-in step expired", "Your two-factor step expired. Sign in again with your password.", "")), nil
	}
	res, err := h.auth.VerifyTwoFactor(ctx, pending, request.Body.Code)
	if err != nil {
		return twoFactorLoginError(err, func(p api.Problem) api.VerifyTwoFactorResponseObject {
			return api.VerifyTwoFactor401ApplicationProblemPlusJSONResponse(p)
		}, func(p api.Problem) api.VerifyTwoFactorResponseObject {
			return api.VerifyTwoFactor429ApplicationProblemPlusJSONResponse(p)
		}), nil
	}
	return api.VerifyTwoFactor200JSONResponse{
		Body:    sessionUser(res.User),
		Headers: api.VerifyTwoFactor200ResponseHeaders{SetCookie: ptr(h.sessionCookie(res.Token, res.Session.ExpiresAt))},
	}, nil
}

func (h *handler) VerifyRecoveryCode(ctx context.Context, request api.VerifyRecoveryCodeRequestObject) (api.VerifyRecoveryCodeResponseObject, error) {
	pending := pendingFrom(ctx)
	if request.Body == nil || pending == "" {
		return api.VerifyRecoveryCode401ApplicationProblemPlusJSONResponse(prob(http.StatusUnauthorized, "Sign-in step expired", "Your two-factor step expired. Sign in again with your password.", "")), nil
	}
	res, err := h.auth.VerifyRecovery(ctx, pending, request.Body.Code)
	if err != nil {
		return twoFactorLoginError(err, func(p api.Problem) api.VerifyRecoveryCodeResponseObject {
			return api.VerifyRecoveryCode401ApplicationProblemPlusJSONResponse(p)
		}, func(p api.Problem) api.VerifyRecoveryCodeResponseObject {
			return api.VerifyRecoveryCode429ApplicationProblemPlusJSONResponse(p)
		}), nil
	}
	return api.VerifyRecoveryCode200JSONResponse{
		Body:    sessionUser(res.User),
		Headers: api.VerifyRecoveryCode200ResponseHeaders{SetCookie: ptr(h.sessionCookie(res.Token, res.Session.ExpiresAt))},
	}, nil
}

func twoFactorLoginError[T any](err error, unauthorized, tooMany func(api.Problem) T) T {
	if errors.Is(err, entity.ErrPendingNotFound) {
		return tooMany(prob(http.StatusTooManyRequests, "Too many attempts", "Too many wrong codes. Sign in again with your password and a fresh code.", ""))
	}
	return unauthorized(prob(http.StatusUnauthorized, "Incorrect code", "That code didn't match. Codes change every 30 seconds — enter the newest one.", ""))
}

func pendingFrom(ctx context.Context) string {
	return entity.PendingTokenFrom(ctx)
}

func (h *handler) ForgotPassword(ctx context.Context, request api.ForgotPasswordRequestObject) (api.ForgotPasswordResponseObject, error) {
	if request.Body == nil {
		return api.ForgotPassword400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	info := entity.RequestInfoFrom(ctx)
	if err := h.auth.RequestPasswordReset(ctx, request.Body.Email, info.IP); err != nil {
		return nil, err
	}
	return api.ForgotPassword202Response{}, nil
}

func (h *handler) ResetPassword(ctx context.Context, request api.ResetPasswordRequestObject) (api.ResetPasswordResponseObject, error) {
	if request.Body == nil {
		return api.ResetPassword400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	err := h.auth.ResetPassword(ctx, request.Body.Token, request.Body.Password)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrPasswordResetNotFound):
			return api.ResetPassword404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Reset link invalid", "This reset link is not valid. Request a new one from the 'Forgot password' page.", "")), nil
		case errors.Is(err, entity.ErrPasswordResetInvalid):
			return api.ResetPassword410ApplicationProblemPlusJSONResponse(prob(http.StatusGone, "Reset link expired", "This reset link has expired or was already used. Request a new one.", "token-expired")), nil
		case isValidation(err):
			return api.ResetPassword400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Weak password", validationDetail(err), "")), nil
		default:
			return nil, err
		}
	}
	return api.ResetPassword204Response{}, nil
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

func (h *handler) PreviewInvite(ctx context.Context, request api.PreviewInviteRequestObject) (api.PreviewInviteResponseObject, error) {
	if request.Body == nil {
		return api.PreviewInvite400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	inv, err := h.auth.InvitePreview(ctx, request.Body.Token)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInviteNotFound), errors.Is(err, entity.ErrInviteRevoked):
			return api.PreviewInvite404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Invitation not found", "This invitation link is not valid.", "")), nil
		case errors.Is(err, entity.ErrInviteExpired), errors.Is(err, entity.ErrInviteAlreadyAccepted):
			return api.PreviewInvite410ApplicationProblemPlusJSONResponse(prob(http.StatusGone, "Invitation unavailable", "This invitation has expired or was already used. Ask an admin to send a new one.", "token-expired")), nil
		default:
			return nil, err
		}
	}
	return api.PreviewInvite200JSONResponse{
		Email:     inv.Email,
		Workspace: inv.WorkspaceName,
		InvitedBy: inv.InvitedByName,
		SentAt:    inv.CreatedAt,
	}, nil
}

func (h *handler) AcceptInvite(ctx context.Context, request api.AcceptInviteRequestObject) (api.AcceptInviteResponseObject, error) {
	if request.Body == nil {
		return api.AcceptInvite400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	b := request.Body
	info := entity.RequestInfoFrom(ctx)
	res, err := h.auth.AcceptInvite(ctx, entity.AcceptInvite{
		Token: b.Token, Name: b.Name, Password: b.Password, Timezone: b.Timezone,
	}, info.IP, info.UserAgent)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInviteNotFound), errors.Is(err, entity.ErrInviteRevoked):
			return api.AcceptInvite404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Invitation not found", "This invitation link is not valid.", "")), nil
		case errors.Is(err, entity.ErrInviteExpired):
			return api.AcceptInvite410ApplicationProblemPlusJSONResponse(prob(http.StatusGone, "Invitation expired", "This invitation has expired. Ask an admin to send a new one.", "token-expired")), nil
		case errors.Is(err, entity.ErrInviteAlreadyAccepted):
			return api.AcceptInvite409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Already accepted", "This invitation was already accepted. Sign in instead.", "")), nil
		case isValidation(err):
			return api.AcceptInvite400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid details", validationDetail(err), "")), nil
		default:
			return nil, err
		}
	}
	return api.AcceptInvite200JSONResponse{
		Body:    sessionUser(res.User),
		Headers: api.AcceptInvite200ResponseHeaders{SetCookie: ptr(h.sessionCookie(res.Token, res.Session.ExpiresAt))},
	}, nil
}

func sessionUser(u entity.User) api.SessionUser {
	return api.SessionUser{Id: u.ID, Name: u.Name, Email: u.Email, Timezone: u.Timezone}
}
