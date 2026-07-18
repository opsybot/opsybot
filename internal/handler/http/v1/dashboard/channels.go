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
	ch, err := h.channels.Add(ctx, entity.NewChannel{Type: entity.ChannelType(request.Body.Type), Detail: request.Body.Detail})
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

func (h *handler) VerifyChannel(ctx context.Context, request api.VerifyChannelRequestObject) (api.VerifyChannelResponseObject, error) {
	err := h.channels.Verify(ctx, request.ChannelId)
	if err != nil {
		if errors.Is(err, entity.ErrChannelNotFound) {
			return api.VerifyChannel404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such channel.", "")), nil
		}
		return nil, err
	}
	return api.VerifyChannel204Response{}, nil
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
