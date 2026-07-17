package dashboard

import (
	"context"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/service"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

type handler struct {
	cfg        config.Auth
	auth       service.Auth
	workspaces service.Workspaces
	members    service.Members
	users      service.Users
	channels   service.Channels
	teams      service.Teams
	apikeys    service.APIKeys
	audits     service.Audits
}

func New(cfg config.Auth, auth service.Auth, workspaces service.Workspaces, members service.Members, users service.Users, channels service.Channels, teams service.Teams, apikeys service.APIKeys, audits service.Audits) api.StrictServerInterface {
	return &handler{cfg: cfg, auth: auth, workspaces: workspaces, members: members, users: users, channels: channels, teams: teams, apikeys: apikeys, audits: audits}
}

func (h *handler) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: api.HealthStatusOk}, nil
}
