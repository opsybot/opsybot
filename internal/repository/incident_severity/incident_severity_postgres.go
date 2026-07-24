package incident_severity

import (
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.IncidentSeverity {
	return &repo{db: db}
}

func (r *repo) List(ctx context.Context, workspaceID string) ([]entity.IncidentSeverity, error) {
	rows, err := dbpostgres.IncidentSeverities(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("position ASC, level ASC"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list severities: %w", err)
	}
	out := make([]entity.IncidentSeverity, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.IncidentSeverity{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			Level:       m.Level,
			Label:       m.Label,
			Definition:  m.Definition,
			Tone:        m.Tone,
			Position:    m.Position,
		})
	}
	return out, nil
}

func (r *repo) Exists(ctx context.Context, workspaceID, level string) (bool, error) {
	exists, err := dbpostgres.IncidentSeverities(
		qm.Where("workspace_id = ? AND level = ?", workspaceID, level),
	).Exists(ctx, r.db.Querier(ctx))
	if err != nil {
		return false, fmt.Errorf("check severity: %w", err)
	}
	return exists, nil
}

func (r *repo) Replace(ctx context.Context, workspaceID string, severities []entity.IncidentSeverity) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.IncidentSeverities(
		qm.Where("workspace_id = ?", workspaceID),
	).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("clear severities: %w", err)
	}
	cols := boil.Whitelist("workspace_id", "level", "label", "definition", "tone", "position")
	for i, s := range severities {
		m := &dbpostgres.IncidentSeverity{
			WorkspaceID: workspaceID,
			Level:       s.Level,
			Label:       s.Label,
			Definition:  s.Definition,
			Tone:        s.Tone,
			Position:    i,
		}
		if err := m.Insert(ctx, exec, cols); err != nil {
			return fmt.Errorf("insert severity: %w", err)
		}
	}
	return nil
}

func (r *repo) SeedDefaults(ctx context.Context, workspaceID string) error {
	exec := r.db.Querier(ctx)
	exists, err := dbpostgres.IncidentSeverities(
		qm.Where("workspace_id = ?", workspaceID),
	).Exists(ctx, exec)
	if err != nil {
		return fmt.Errorf("check severities: %w", err)
	}
	if exists {
		return nil
	}
	cols := boil.Whitelist("workspace_id", "level", "label", "definition", "tone", "position")
	for _, s := range entity.DefaultSeverities() {
		m := &dbpostgres.IncidentSeverity{
			WorkspaceID: workspaceID,
			Level:       s.Level,
			Label:       s.Label,
			Definition:  s.Definition,
			Tone:        s.Tone,
			Position:    s.Position,
		}
		if err := m.Insert(ctx, exec, cols); err != nil {
			if _, ok := postgres.UniqueViolation(err); ok {
				continue
			}
			return fmt.Errorf("seed severity: %w", err)
		}
	}
	return nil
}
