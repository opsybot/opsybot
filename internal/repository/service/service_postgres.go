package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
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

func New(db postgres.Client) repository.Service {
	return &repo{db: db}
}

func toEntity(m *dbpostgres.Service) entity.Service {
	return entity.Service{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		Slug:        m.Slug,
		Name:        m.Name,
		TeamID:      m.TeamID.String,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (r *repo) List(ctx context.Context, workspaceID string) ([]entity.Service, error) {
	rows, err := dbpostgres.Services(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("name ASC, id ASC"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	out := make([]entity.Service, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEntity(m))
	}
	return out, nil
}

func (r *repo) GetByID(ctx context.Context, workspaceID, id string) (entity.Service, error) {
	m, err := dbpostgres.Services(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
	).One(ctx, r.db.Querier(ctx))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Service{}, entity.ErrServiceNotFound
	}
	if err != nil {
		return entity.Service{}, fmt.Errorf("get service: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) SlugExists(ctx context.Context, workspaceID, slug string) (bool, error) {
	exists, err := dbpostgres.Services(
		qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug),
	).Exists(ctx, r.db.Querier(ctx))
	if err != nil {
		return false, fmt.Errorf("check service slug: %w", err)
	}
	return exists, nil
}

func (r *repo) ExistingIDs(ctx context.Context, workspaceID string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := dbpostgres.Services(
		qm.Select("id"),
		qm.Where("workspace_id = ?", workspaceID),
		qm.WhereIn("id IN ?", anySlice(ids)...),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("check service ids: %w", err)
	}
	out := make([]string, 0, len(rows))
	for _, m := range rows {
		out = append(out, m.ID)
	}
	return out, nil
}

func (r *repo) Create(ctx context.Context, s entity.Service) (entity.Service, error) {
	m := &dbpostgres.Service{
		WorkspaceID: s.WorkspaceID,
		Slug:        s.Slug,
		Name:        s.Name,
		Description: s.Description,
	}
	if s.TeamID != "" {
		m.TeamID = null.StringFrom(s.TeamID)
	}
	cols := boil.Whitelist("workspace_id", "slug", "name", "team_id", "description")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.Service{}, entity.ErrServiceSlugTaken
		}
		return entity.Service{}, fmt.Errorf("create service: %w", err)
	}
	return r.GetByID(ctx, s.WorkspaceID, m.ID)
}

func (r *repo) Update(ctx context.Context, workspaceID, id, name, teamID, description string) (entity.Service, error) {
	values := dbpostgres.M{
		"name":        name,
		"description": description,
		"updated_at":  time.Now().UTC(),
	}
	if teamID != "" {
		values["team_id"] = teamID
	} else {
		values["team_id"] = nil
	}
	affected, err := dbpostgres.Services(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
	).UpdateAll(ctx, r.db.Querier(ctx), values)
	if err != nil {
		return entity.Service{}, fmt.Errorf("update service: %w", err)
	}
	if affected == 0 {
		return entity.Service{}, entity.ErrServiceNotFound
	}
	return r.GetByID(ctx, workspaceID, id)
}

func (r *repo) Delete(ctx context.Context, workspaceID, id string) error {
	affected, err := dbpostgres.Services(
		qm.Where("workspace_id = ? AND id = ?", workspaceID, id),
	).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	if affected == 0 {
		return entity.ErrServiceNotFound
	}
	return nil
}

func anySlice(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}
