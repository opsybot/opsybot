package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) ListKeys(ctx context.Context, request api.ListKeysRequestObject) (api.ListKeysResponseObject, error) {
	list, err := h.apikeys.List(ctx, request.WorkspaceId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ListKeys403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "You don't have permission to view API keys.", "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ListKeys404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		default:
			return nil, err
		}
	}
	return api.ListKeys200JSONResponse{Personal: keyDTOs(list.Personal), Workspace: keyDTOs(list.Workspace)}, nil
}

func (h *handler) CreateKey(ctx context.Context, request api.CreateKeyRequestObject) (api.CreateKeyResponseObject, error) {
	if request.Body == nil {
		return api.CreateKey400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	key, secret, err := h.apikeys.Create(ctx, request.WorkspaceId, entity.NewAPIKey{
		Name:   request.Body.Name,
		Kind:   entity.KeyKind(request.Body.Kind),
		Scopes: toScopes(request.Body.Scopes),
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.CreateKey403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "You don't have permission to create that kind of key.", "")), nil
		case errors.Is(err, entity.ErrAPIKeyInvalidName), errors.Is(err, entity.ErrAPIKeyInvalidKind), errors.Is(err, entity.ErrAPIKeyInvalidScope):
			return api.CreateKey400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid key", keyValidationDetail(err), "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.CreateKey404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		default:
			return nil, err
		}
	}
	return api.CreateKey201JSONResponse{Key: keyDTO(key), Secret: secret}, nil
}

func (h *handler) RevokeKey(ctx context.Context, request api.RevokeKeyRequestObject) (api.RevokeKeyResponseObject, error) {
	err := h.apikeys.Revoke(ctx, request.WorkspaceId, request.KeyId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.RevokeKey403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "You don't have permission to revoke that key.", "")), nil
		case errors.Is(err, entity.ErrAPIKeyNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.RevokeKey404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such API key.", "")), nil
		case errors.Is(err, entity.ErrAPIKeyRevoked):
			return api.RevokeKey409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Already revoked", "That key is already revoked.", "")), nil
		default:
			return nil, err
		}
	}
	return api.RevokeKey204Response{}, nil
}

func keyValidationDetail(err error) string {
	switch {
	case errors.Is(err, entity.ErrAPIKeyInvalidName):
		return "Give the key a name of 60 characters or fewer."
	case errors.Is(err, entity.ErrAPIKeyInvalidKind):
		return "Key kind must be personal or workspace."
	default:
		return "Pick at least one valid scope."
	}
}

func toScopes(in []api.Scope) []entity.Scope {
	out := make([]entity.Scope, 0, len(in))
	for _, s := range in {
		out = append(out, entity.Scope(s))
	}
	return out
}

func fromScopes(in []entity.Scope) []api.Scope {
	out := make([]api.Scope, 0, len(in))
	for _, s := range in {
		out = append(out, api.Scope(s))
	}
	return out
}

func keyDTO(k entity.APIKey) api.ApiKey {
	dto := api.ApiKey{
		Id:        k.ID,
		Name:      k.Name,
		Kind:      api.ApiKeyKind(k.Kind),
		Scopes:    fromScopes(k.Scopes),
		Hint:      k.TokenHint,
		CreatedAt: k.CreatedAt,
	}
	if !k.LastUsedAt.IsZero() {
		dto.LastUsedAt = &k.LastUsedAt
	}
	return dto
}

func keyDTOs(in []entity.APIKey) []api.ApiKey {
	out := make([]api.ApiKey, 0, len(in))
	for _, k := range in {
		out = append(out, keyDTO(k))
	}
	return out
}
