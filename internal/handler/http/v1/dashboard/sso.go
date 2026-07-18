package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) GetSsoConfig(ctx context.Context, request api.GetSsoConfigRequestObject) (api.GetSsoConfigResponseObject, error) {
	conn, err := h.sso.GetConfig(ctx, request.WorkspaceId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.GetSsoConfig403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can view SSO settings.", "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.GetSsoConfig404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		default:
			return nil, err
		}
	}
	return api.GetSsoConfig200JSONResponse(ssoDTO(conn)), nil
}

func (h *handler) PutSsoConfig(ctx context.Context, request api.PutSsoConfigRequestObject) (api.PutSsoConfigResponseObject, error) {
	if request.Body == nil {
		return api.PutSsoConfig400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	conn, err := h.sso.SaveConfig(ctx, request.WorkspaceId, entity.SSOConfigInput{
		Mode:                entity.SSOMode(request.Body.Mode),
		Issuer:              strParam(request.Body.Issuer),
		ClientID:            strParam(request.Body.ClientId),
		ClientSecret:        strParam(request.Body.ClientSecret),
		ClearClientSecret:   boolParam(request.Body.ClearClientSecret),
		Scopes:              strsParam(request.Body.Scopes),
		SAMLMetadataURL:     strParam(request.Body.SamlMetadataUrl),
		Enabled:             request.Body.Enabled,
		Required:            request.Body.Required,
		JITProvisioning:     request.Body.JitProvisioning,
		AllowedEmailDomains: strsParam(request.Body.AllowedEmailDomains),
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.PutSsoConfig403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can change SSO settings.", "")), nil
		case isValidation(err):
			return api.PutSsoConfig400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid SSO settings", validationDetail(err), "")), nil
		case errors.Is(err, entity.ErrSSOUnavailable):
			return api.PutSsoConfig409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Secret storage unavailable", "This instance has no auth secret key configured, so the client secret can't be stored. Ask an admin to set OPSYBOT_AUTH_SECRET_KEY.", "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.PutSsoConfig404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		default:
			return nil, err
		}
	}
	return api.PutSsoConfig200JSONResponse(ssoDTO(conn)), nil
}

func ssoDTO(c entity.SSOConnection) api.SsoConfig {
	return api.SsoConfig{
		Mode:                api.SsoMode(c.Mode),
		Issuer:              c.Issuer,
		ClientId:            c.ClientID,
		HasClientSecret:     c.HasClientSecret,
		Scopes:              nonNilStrings(c.Scopes),
		SamlMetadataUrl:     c.SAMLMetadataURL,
		Enabled:             c.Enabled,
		Required:            c.Required,
		JitProvisioning:     c.JITProvisioning,
		AllowedEmailDomains: nonNilStrings(c.AllowedEmailDomains),
	}
}

func boolParam(p *bool) bool {
	return p != nil && *p
}

func strsParam(p *[]string) []string {
	if p == nil {
		return nil
	}
	return *p
}

func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
