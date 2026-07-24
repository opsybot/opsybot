package service

//go:generate go tool mockgen -source=incidents.go -destination=./incidents/incidents_mock.go -package=incidents

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type Incidents interface {
	List(ctx context.Context, workspaceSlug string, filter entity.IncidentFilter) (entity.IncidentPage, error)
	Get(ctx context.Context, workspaceSlug, id string) (entity.Incident, error)
	Declare(ctx context.Context, workspaceSlug string, in entity.IncidentDeclare) (entity.Incident, error)
	Update(ctx context.Context, workspaceSlug, id string, in entity.IncidentUpdate) (entity.Incident, error)
	ChangeStatus(ctx context.Context, workspaceSlug, id string, to entity.IncidentStatus) (entity.Incident, error)
	ChangeSeverity(ctx context.Context, workspaceSlug, id, level string) (entity.Incident, error)
	Resolve(ctx context.Context, workspaceSlug, id, summary string) (entity.Incident, error)
	Reopen(ctx context.Context, workspaceSlug, id string) (entity.Incident, error)
	SetCustomFields(ctx context.Context, workspaceSlug, id string, fields map[string]string) (entity.Incident, error)
	LinkAlert(ctx context.Context, workspaceSlug, id, alertID string) (entity.Incident, error)
	UnlinkAlert(ctx context.Context, workspaceSlug, id, alertID string) (entity.Incident, error)
	Relate(ctx context.Context, workspaceSlug, id, relatedID string, kind entity.IncidentRelationKind) (entity.Incident, error)
	Unrelate(ctx context.Context, workspaceSlug, id, relationID string) (entity.Incident, error)
	AddFollowup(ctx context.Context, workspaceSlug, id string, in entity.NewFollowup) (entity.Incident, error)
	ToggleFollowup(ctx context.Context, workspaceSlug, id, followupID string, done bool) (entity.Incident, error)
	ListSeverities(ctx context.Context, workspaceSlug string) ([]entity.IncidentSeverity, error)
	SaveSeverities(ctx context.Context, workspaceSlug string, severities []entity.IncidentSeverity) ([]entity.IncidentSeverity, error)
	ListFieldDefs(ctx context.Context, workspaceSlug string) ([]entity.IncidentFieldDef, error)
	SaveFieldDefs(ctx context.Context, workspaceSlug string, defs []entity.IncidentFieldDef) ([]entity.IncidentFieldDef, error)
}
