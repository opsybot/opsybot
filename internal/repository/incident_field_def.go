package repository

//go:generate go tool mockgen -source=incident_field_def.go -destination=./incident_field_def/incident_field_def_mock.go -package=incident_field_def

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type IncidentFieldDef interface {
	List(ctx context.Context, workspaceID string) ([]entity.IncidentFieldDef, error)
	Replace(ctx context.Context, workspaceID string, defs []entity.IncidentFieldDef) error
}
