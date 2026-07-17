package dashboard

import (
	"context"

	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

type handler struct{}

func New() api.StrictServerInterface {
	return &handler{}
}

func (h *handler) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: api.Ok}, nil
}
