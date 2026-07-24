package incident_field_def

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.IncidentFieldDef {
	return &repo{db: db}
}

func optionsToJSON(options []string) (types.JSON, error) {
	if options == nil {
		options = []string{}
	}
	raw, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("marshal field options: %w", err)
	}
	return types.JSON(raw), nil
}

func optionsFromJSON(raw types.JSON) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	return out
}

func (r *repo) List(ctx context.Context, workspaceID string) ([]entity.IncidentFieldDef, error) {
	rows, err := dbpostgres.IncidentFieldDefs(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("position ASC, name ASC"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list field defs: %w", err)
	}
	out := make([]entity.IncidentFieldDef, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.IncidentFieldDef{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			Slug:        m.Slug,
			Name:        m.Name,
			Kind:        entity.CustomFieldKind(m.Kind),
			Options:     optionsFromJSON(m.Options),
			Position:    m.Position,
		})
	}
	return out, nil
}

func (r *repo) Replace(ctx context.Context, workspaceID string, defs []entity.IncidentFieldDef) error {
	exec := r.db.Querier(ctx)
	keep := make([]any, 0, len(defs))
	for _, d := range defs {
		if d.ID != "" {
			keep = append(keep, d.ID)
		}
	}
	delMods := []qm.QueryMod{qm.Where("workspace_id = ?", workspaceID)}
	if len(keep) > 0 {
		delMods = append(delMods, qm.WhereNotIn("id NOT IN ?", keep...))
	}
	if _, err := dbpostgres.IncidentFieldDefs(delMods...).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("prune field defs: %w", err)
	}
	for i, d := range defs {
		options, err := optionsToJSON(d.Options)
		if err != nil {
			return err
		}
		if d.ID != "" {
			values := dbpostgres.M{
				"slug":       d.Slug,
				"name":       d.Name,
				"kind":       string(d.Kind),
				"options":    options,
				"position":   i,
				"updated_at": time.Now().UTC(),
			}
			affected, err := dbpostgres.IncidentFieldDefs(
				qm.Where("workspace_id = ? AND id = ?", workspaceID, d.ID),
			).UpdateAll(ctx, exec, values)
			if err != nil {
				if _, ok := postgres.UniqueViolation(err); ok {
					return entity.ErrFieldSlugTaken
				}
				return fmt.Errorf("update field def: %w", err)
			}
			if affected > 0 {
				continue
			}
		}
		m := &dbpostgres.IncidentFieldDef{
			WorkspaceID: workspaceID,
			Slug:        d.Slug,
			Name:        d.Name,
			Kind:        string(d.Kind),
			Options:     options,
			Position:    i,
		}
		cols := boil.Whitelist("workspace_id", "slug", "name", "kind", "options", "position")
		if err := m.Insert(ctx, exec, cols); err != nil {
			if _, ok := postgres.UniqueViolation(err); ok {
				return entity.ErrFieldSlugTaken
			}
			return fmt.Errorf("insert field def: %w", err)
		}
	}
	return nil
}
