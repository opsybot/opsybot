package workspace

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

const selectColumnsW = `w.id, w.slug, w.name, w.timezone, w.environment, w.created_by, w.created_at`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Workspace {
	return &repo{db: db}
}

func scanWorkspace(row interface {
	Scan(dest ...any) error
}) (dbpostgres.Workspace, error) {
	var m dbpostgres.Workspace
	err := row.Scan(&m.ID, &m.Slug, &m.Name, &m.Timezone, &m.Environment, &m.CreatedBy, &m.CreatedAt)
	return m, err
}

func toEntity(m dbpostgres.Workspace) entity.Workspace {
	return entity.Workspace{
		ID:          m.ID,
		Slug:        m.Slug,
		Name:        m.Name,
		Timezone:    m.Timezone,
		Environment: m.Environment,
		CreatedBy:   m.CreatedBy.String,
		CreatedAt:   m.CreatedAt,
	}
}

func (r *repo) Create(ctx context.Context, w entity.NewWorkspace, createdBy string) (entity.Workspace, error) {
	m := &dbpostgres.Workspace{
		Slug:        w.Slug,
		Name:        w.Name,
		Timezone:    w.Timezone,
		Environment: w.Environment,
		CreatedBy:   null.StringFrom(createdBy),
	}
	if err := m.Insert(ctx, r.db.Querier(ctx),
		boil.Whitelist("slug", "name", "timezone", "environment", "created_by")); err != nil {
		if name, ok := postgres.UniqueViolation(err); ok && name == "workspaces_slug_uq" {
			return entity.Workspace{}, entity.ErrWorkspaceSlugTaken
		}
		return entity.Workspace{}, fmt.Errorf("create workspace: %w", err)
	}
	return toEntity(*m), nil
}

func (r *repo) GetByID(ctx context.Context, id string) (entity.Workspace, error) {
	m, err := dbpostgres.FindWorkspace(ctx, r.db.Querier(ctx), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Workspace{}, entity.ErrWorkspaceNotFound
		}
		return entity.Workspace{}, fmt.Errorf("get workspace by id: %w", err)
	}
	return toEntity(*m), nil
}

func (r *repo) GetBySlug(ctx context.Context, slug string) (entity.Workspace, error) {
	m, err := dbpostgres.Workspaces(qm.Where("slug = ?", slug)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Workspace{}, entity.ErrWorkspaceNotFound
		}
		return entity.Workspace{}, fmt.Errorf("get workspace by slug: %w", err)
	}
	return toEntity(*m), nil
}

func (r *repo) Update(ctx context.Context, id string, u entity.WorkspaceUpdate) error {
	n, err := dbpostgres.Workspaces(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"name": u.Name, "timezone": u.Timezone, "environment": u.Environment, "updated_at": time.Now()})
	if err != nil {
		return fmt.Errorf("update workspace: %w", err)
	}
	if n == 0 {
		return entity.ErrWorkspaceNotFound
	}
	return nil
}

func (r *repo) ListActiveByUser(ctx context.Context, userID string) ([]entity.Workspace, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT `+selectColumnsW+`
		 FROM workspaces w
		 JOIN workspace_members m ON m.workspace_id = w.id
		 WHERE m.user_id = $1 AND m.status = 'active'
		 ORDER BY w.name`, userID)
	if err != nil {
		return nil, fmt.Errorf("list workspaces by user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []entity.Workspace
	for rows.Next() {
		m, err := scanWorkspace(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workspace: %w", err)
		}
		out = append(out, toEntity(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workspaces: %w", err)
	}
	return out, nil
}
