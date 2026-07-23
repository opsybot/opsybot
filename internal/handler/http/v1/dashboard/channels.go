package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) ListChannels(ctx context.Context, _ api.ListChannelsRequestObject) (api.ListChannelsResponseObject, error) {
	list, err := h.channels.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]api.Channel, 0, len(list))
	for _, c := range list {
		items = append(items, channelDTO(c))
	}
	return api.ListChannels200JSONResponse{Items: items}, nil
}

func (h *handler) CreateChannel(ctx context.Context, request api.CreateChannelRequestObject) (api.CreateChannelResponseObject, error) {
	if request.Body == nil {
		return api.CreateChannel400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	in := entity.NewChannel{Type: entity.ChannelType(request.Body.Type), Detail: request.Body.Detail}
	if request.Body.Label != nil {
		in.Label = *request.Body.Label
	}
	if request.Body.Secret != nil {
		in.Secret = *request.Body.Secret
	}
	ch, err := h.channels.Add(ctx, in)
	if err != nil {
		switch {
		case isValidation(err):
			return api.CreateChannel400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid channel", validationDetail(err), "")), nil
		case errors.Is(err, entity.ErrChannelDuplicate):
			return api.CreateChannel409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Already added", "You've already added that channel.", "")), nil
		default:
			return nil, err
		}
	}
	return api.CreateChannel201JSONResponse(channelDTO(ch)), nil
}

func (h *handler) StartChannelVerification(ctx context.Context, request api.StartChannelVerificationRequestObject) (api.StartChannelVerificationResponseObject, error) {
	v, err := h.channels.StartVerification(ctx, request.ChannelId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUnauthenticated):
			return api.StartChannelVerification401ApplicationProblemPlusJSONResponse(prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")), nil
		case errors.Is(err, entity.ErrChannelNotFound):
			return api.StartChannelVerification404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such channel.", "")), nil
		default:
			return nil, err
		}
	}
	out := api.ChannelVerification{Method: api.ChannelVerificationMethod(v.Method), ExpiresAt: v.ExpiresAt}
	if v.DeepLink != "" {
		out.DeepLink = ptr(v.DeepLink)
	}
	if v.Detail != "" {
		out.Detail = ptr(v.Detail)
	}
	return api.StartChannelVerification200JSONResponse(out), nil
}

func (h *handler) VerifyChannel(ctx context.Context, request api.VerifyChannelRequestObject) (api.VerifyChannelResponseObject, error) {
	code := ""
	if request.Body != nil && request.Body.Code != nil {
		code = *request.Body.Code
	}
	err := h.channels.CompleteVerification(ctx, request.ChannelId, code)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrChannelVerifyInvalid), errors.Is(err, entity.ErrChannelVerifyExpired):
			return api.VerifyChannel400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid code", "That code is wrong or expired. Start verification again.", "")), nil
		case errors.Is(err, entity.ErrUnauthenticated):
			return api.VerifyChannel401ApplicationProblemPlusJSONResponse(prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")), nil
		case errors.Is(err, entity.ErrChannelNotFound):
			return api.VerifyChannel404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such channel.", "")), nil
		default:
			return nil, err
		}
	}
	return api.VerifyChannel204Response{}, nil
}

func (h *handler) TestChannel(ctx context.Context, request api.TestChannelRequestObject) (api.TestChannelResponseObject, error) {
	result, err := h.channels.SendTest(ctx, request.ChannelId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUnauthenticated):
			return api.TestChannel401ApplicationProblemPlusJSONResponse(prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")), nil
		case errors.Is(err, entity.ErrChannelNotFound):
			return api.TestChannel404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such channel.", "")), nil
		case errors.Is(err, entity.ErrChannelNotVerified):
			return api.TestChannel409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Not verified", "Verify this channel before sending a test.", "")), nil
		case errors.Is(err, entity.ErrRateLimited):
			return api.TestChannel429ApplicationProblemPlusJSONResponse(prob(http.StatusTooManyRequests, "Slow down", "You've sent too many test notifications. Try again later.", "")), nil
		default:
			return nil, err
		}
	}
	return api.TestChannel200JSONResponse{Delivered: result.Delivered, Detail: result.Detail}, nil
}

func (h *handler) DeleteChannel(ctx context.Context, request api.DeleteChannelRequestObject) (api.DeleteChannelResponseObject, error) {
	err := h.channels.Remove(ctx, request.ChannelId)
	if err != nil {
		if errors.Is(err, entity.ErrChannelNotFound) {
			return api.DeleteChannel404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such channel.", "")), nil
		}
		return nil, err
	}
	return api.DeleteChannel204Response{}, nil
}

func channelDTO(c entity.Channel) api.Channel {
	return api.Channel{Id: c.ID, Type: api.ChannelType(c.Type), Detail: c.Detail, Verified: c.Verified}
}
