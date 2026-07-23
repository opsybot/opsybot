package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func chatProblem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden, prob(http.StatusForbidden, "Forbidden", "You do not have access to chat connections in this workspace.", "")
	case errors.Is(err, entity.ErrUnauthenticated):
		return http.StatusUnauthorized, prob(http.StatusUnauthorized, "Unauthenticated", "Sign in to continue.", "")
	case errors.Is(err, entity.ErrChatConnectionInvalid):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Connection rejected", "The provider rejected the connection. Check the workspace details and try again.", "")
	case errors.Is(err, entity.ErrChatProviderNotConfigured):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Provider not configured", "This chat provider is not configured on the server yet. An admin needs to set its app credentials in the environment.", "")
	case errors.Is(err, entity.ErrChatOAuthUnsupported):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Not available", "This provider does not use the OAuth install flow.", "")
	case errors.Is(err, entity.ErrChatOAuthExchange), errors.Is(err, entity.ErrChatOAuthStateInvalid):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Install failed", "The provider install could not be completed. Start the connection again.", "")
	case errors.Is(err, entity.ErrChatSecretUnavailable):
		return http.StatusBadRequest, prob(http.StatusBadRequest, "Secret storage unavailable", "This instance has no auth secret key configured, so the bot token can't be stored. Ask an admin to set OPSYBOT_AUTH_SECRET_KEY.", "")
	case errors.Is(err, entity.ErrChatNotConnected):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not connected", "That chat provider is not connected.", "")
	case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
		return http.StatusNotFound, prob(http.StatusNotFound, "Not found", "That workspace does not exist.", "")
	default:
		return 0, api.Problem{}
	}
}

func chatConnectionDTO(c entity.ChatConnection) api.ChatConnection {
	dto := api.ChatConnection{
		Provider:         api.ChatConnectionProvider(c.Provider),
		ExternalName:     c.ExternalName,
		Health:           api.ChatConnectionHealth(c.Health),
		HealthNote:       c.HealthNote,
		Enabled:          c.Enabled,
		NamingPattern:    c.NamingPattern,
		AnnounceChannel:  c.AnnounceChannel,
		ArchiveOnResolve: c.ArchiveOnResolve,
		Linked:           c.Linked,
	}
	if c.LinkedHandle != "" {
		handle := c.LinkedHandle
		dto.LinkedHandle = &handle
	}
	if !c.HealthCheckedAt.IsZero() {
		at := c.HealthCheckedAt
		dto.HealthCheckedAt = &at
	}
	return dto
}

func (h *handler) ListChatConnections(ctx context.Context, request api.ListChatConnectionsRequestObject) (api.ListChatConnectionsResponseObject, error) {
	list, err := h.chats.List(ctx, request.WorkspaceId)
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.ListChatConnections401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ListChatConnections403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ListChatConnections404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	items := make([]api.ChatConnection, 0, len(list))
	for _, c := range list {
		items = append(items, chatConnectionDTO(c))
	}
	return api.ListChatConnections200JSONResponse{Items: items}, nil
}

func (h *handler) ConnectChat(ctx context.Context, request api.ConnectChatRequestObject) (api.ConnectChatResponseObject, error) {
	if request.Body == nil {
		return api.ConnectChat400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	in := entity.ChatConnectInput{Provider: entity.ChatProvider(request.Body.Provider)}
	if request.Body.BotToken != nil {
		in.BotToken = *request.Body.BotToken
	}
	if request.Body.ExternalId != nil {
		in.ExternalID = *request.Body.ExternalId
	}
	conn, err := h.chats.Connect(ctx, request.WorkspaceId, in)
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.ConnectChat400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.ConnectChat401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.ConnectChat403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.ConnectChat404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	return api.ConnectChat201JSONResponse(chatConnectionDTO(conn)), nil
}

func (h *handler) DeleteChatConnection(ctx context.Context, request api.DeleteChatConnectionRequestObject) (api.DeleteChatConnectionResponseObject, error) {
	err := h.chats.Delete(ctx, request.WorkspaceId, entity.ChatProvider(request.Provider))
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.DeleteChatConnection401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.DeleteChatConnection403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.DeleteChatConnection404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	return api.DeleteChatConnection204Response{}, nil
}

func (h *handler) PutChatDefaults(ctx context.Context, request api.PutChatDefaultsRequestObject) (api.PutChatDefaultsResponseObject, error) {
	if request.Body == nil {
		return api.PutChatDefaults401ApplicationProblemPlusJSONResponse(prob(http.StatusUnauthorized, "Invalid request", "The request body was empty.", "")), nil
	}
	err := h.chats.SetDefaults(ctx, request.WorkspaceId, entity.ChatProvider(request.Provider),
		request.Body.NamingPattern, request.Body.AnnounceChannel, request.Body.ArchiveOnResolve)
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.PutChatDefaults401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.PutChatDefaults403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.PutChatDefaults404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	return api.PutChatDefaults204Response{}, nil
}

func (h *handler) TestChatConnection(ctx context.Context, request api.TestChatConnectionRequestObject) (api.TestChatConnectionResponseObject, error) {
	result, err := h.chats.TestConnection(ctx, request.WorkspaceId, entity.ChatProvider(request.Provider))
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusUnauthorized:
			return api.TestChatConnection401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.TestChatConnection403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.TestChatConnection404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	detail := result.Result.Detail
	if detail == "" {
		detail = "Test message sent."
	}
	return api.TestChatConnection200JSONResponse{Delivered: result.Result.Delivered, Detail: detail}, nil
}

func (h *handler) StartTelegramLink(ctx context.Context, request api.StartTelegramLinkRequestObject) (api.StartTelegramLinkResponseObject, error) {
	url, err := h.chats.StartTelegramLink(ctx, request.WorkspaceId)
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.StartTelegramLink400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.StartTelegramLink401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.StartTelegramLink403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.StartTelegramLink404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	return api.StartTelegramLink200JSONResponse{AuthorizeUrl: url}, nil
}

func (h *handler) StartChatIdentityOAuth(ctx context.Context, request api.StartChatIdentityOAuthRequestObject) (api.StartChatIdentityOAuthResponseObject, error) {
	url, err := h.chats.StartIdentityOAuth(ctx, request.WorkspaceId, entity.ChatProvider(request.Provider))
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.StartChatIdentityOAuth400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.StartChatIdentityOAuth401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.StartChatIdentityOAuth403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.StartChatIdentityOAuth404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	return api.StartChatIdentityOAuth200JSONResponse{AuthorizeUrl: url}, nil
}

func (h *handler) StartChatOAuth(ctx context.Context, request api.StartChatOAuthRequestObject) (api.StartChatOAuthResponseObject, error) {
	url, err := h.chats.StartOAuth(ctx, request.WorkspaceId, entity.ChatProvider(request.Provider))
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.StartChatOAuth400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.StartChatOAuth401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.StartChatOAuth403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.StartChatOAuth404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	return api.StartChatOAuth200JSONResponse{AuthorizeUrl: url}, nil
}

func (h *handler) LinkChatIdentity(ctx context.Context, request api.LinkChatIdentityRequestObject) (api.LinkChatIdentityResponseObject, error) {
	ident, err := h.chats.LinkIdentity(ctx, request.WorkspaceId, entity.ChatProvider(request.Provider))
	if err != nil {
		status, p := chatProblem(err)
		switch status {
		case http.StatusBadRequest:
			return api.LinkChatIdentity400ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusUnauthorized:
			return api.LinkChatIdentity401ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusForbidden:
			return api.LinkChatIdentity403ApplicationProblemPlusJSONResponse(p), nil
		case http.StatusNotFound:
			return api.LinkChatIdentity404ApplicationProblemPlusJSONResponse(p), nil
		default:
			return nil, err
		}
	}
	return api.LinkChatIdentity200JSONResponse{ProviderHandle: ident.ProviderHandle, Verified: ident.Verified}, nil
}
