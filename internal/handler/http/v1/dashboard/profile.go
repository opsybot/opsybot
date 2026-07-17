package dashboard

import (
	"context"

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
