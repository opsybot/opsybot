package dashboard

import (
	"context"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/service"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

type handler struct {
	cfg           config.Auth
	auth          service.Auth
	workspaces    service.Workspaces
	members       service.Members
	users         service.Users
	channels      service.Channels
	teams         service.Teams
	schedules     service.Schedules
	apikeys       service.APIKeys
	audits        service.Audits
	sso           service.SSO
	alerts        service.Alerts
	sources       service.AlertSources
	routes        service.AlertRoutes
	silences      service.Silences
	monitors      service.AlertMonitors
	ingestBaseURL string
}

func New(cfg config.Auth, auth service.Auth, workspaces service.Workspaces, members service.Members, users service.Users, channels service.Channels, teams service.Teams, schedules service.Schedules, apikeys service.APIKeys, audits service.Audits, sso service.SSO, alerts service.Alerts, sources service.AlertSources, routes service.AlertRoutes, silences service.Silences, monitors service.AlertMonitors, cfgIngest config.Ingest) api.StrictServerInterface {
	return &handler{cfg: cfg, auth: auth, workspaces: workspaces, members: members, users: users, channels: channels, teams: teams, schedules: schedules, apikeys: apikeys, audits: audits, sso: sso, alerts: alerts, sources: sources, routes: routes, silences: silences, monitors: monitors, ingestBaseURL: cfgIngest.BaseURL}
}

func (h *handler) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: api.HealthStatusOk}, nil
}
