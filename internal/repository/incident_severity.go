package repository

//go:generate go tool mockgen -source=incident_severity.go -destination=./incident_severity/incident_severity_mock.go -package=incident_severity

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type IncidentSeverity interface {
	List(ctx context.Context, workspaceID string) ([]entity.IncidentSeverity, error)
	Exists(ctx context.Context, workspaceID, level string) (bool, error)
	Replace(ctx context.Context, workspaceID string, severities []entity.IncidentSeverity) error
	SeedDefaults(ctx context.Context, workspaceID string) error
}
