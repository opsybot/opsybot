package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) EnrollTwoFactor(ctx context.Context, _ api.EnrollTwoFactorRequestObject) (api.EnrollTwoFactorResponseObject, error) {
	enrollment, err := h.users.BeginTOTP(ctx)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrTOTPAlreadySetUp):
			return api.EnrollTwoFactor409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Already enabled", "Two-factor is already enabled. Disable it first to re-enroll.", "")), nil
		case errors.Is(err, entity.ErrTOTPUnavailable):
			return api.EnrollTwoFactor409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Two-factor unavailable", "This instance has no auth secret key configured, so authenticator secrets can't be stored. Ask an admin to set OPSYBOT_AUTH_SECRET_KEY.", "")), nil
		default:
			return nil, err
		}
	}
	return api.EnrollTwoFactor200JSONResponse{Secret: enrollment.Secret, OtpauthUri: enrollment.OTPAuthURI}, nil
}

func (h *handler) ActivateTwoFactor(ctx context.Context, request api.ActivateTwoFactorRequestObject) (api.ActivateTwoFactorResponseObject, error) {
	if request.Body == nil {
		return api.ActivateTwoFactor400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	codes, err := h.users.ConfirmTOTP(ctx, request.Body.Code)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrTOTPInvalidCode):
			return api.ActivateTwoFactor400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Incorrect code", "That code didn't match. Rescan the QR code or check your device clock, then enter the newest code.", "")), nil
		case errors.Is(err, entity.ErrTOTPNotEnrolled):
			return api.ActivateTwoFactor409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Not enrolled", "Start two-factor enrollment first.", "")), nil
		default:
			return nil, err
		}
	}
	return api.ActivateTwoFactor200JSONResponse{Codes: codes}, nil
}

func (h *handler) RegenerateRecoveryCodes(ctx context.Context, request api.RegenerateRecoveryCodesRequestObject) (api.RegenerateRecoveryCodesResponseObject, error) {
	if request.Body == nil {
		return api.RegenerateRecoveryCodes400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	codes, err := h.users.RegenerateRecoveryCodes(ctx, request.Body.Code)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrTOTPInvalidCode):
			return api.RegenerateRecoveryCodes400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Incorrect code", "That code didn't match. Enter the newest code from your authenticator.", "")), nil
		case errors.Is(err, entity.ErrTOTPNotEnrolled):
			return api.RegenerateRecoveryCodes409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Not enabled", "Two-factor is not enabled.", "")), nil
		default:
			return nil, err
		}
	}
	return api.RegenerateRecoveryCodes200JSONResponse{Codes: codes}, nil
}

func (h *handler) DisableTwoFactor(ctx context.Context, request api.DisableTwoFactorRequestObject) (api.DisableTwoFactorResponseObject, error) {
	if request.Body == nil {
		return api.DisableTwoFactor400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	err := h.users.DisableTOTP(ctx, request.Body.Code)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrTOTPInvalidCode):
			return api.DisableTwoFactor400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Incorrect code", "That code didn't match. Enter the newest code from your authenticator.", "")), nil
		case errors.Is(err, entity.ErrTOTPNotEnrolled):
			return api.DisableTwoFactor409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Not enabled", "Two-factor is not enabled.", "")), nil
		default:
			return nil, err
		}
	}
	return api.DisableTwoFactor204Response{}, nil
}
