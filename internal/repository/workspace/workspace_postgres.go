package workspace

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const selectColumns = `id, slug, name, timezone, environment, created_by, created_at`

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
	m, err := scanWorkspace(r.db.Querier(ctx).QueryRowContext(ctx,
		`INSERT INTO workspaces (slug, name, timezone, environment, created_by)
		 VALUES ($1, $2, $3, $4, $5) RETURNING `+selectColumns,
		w.Slug, w.Name, w.Timezone, w.Environment, createdBy))
	if err != nil {
		if name, ok := postgres.UniqueViolation(err); ok && name == "workspaces_slug_uq" {
			return entity.Workspace{}, entity.ErrWorkspaceSlugTaken
		}
		return entity.Workspace{}, fmt.Errorf("create workspace: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetByID(ctx context.Context, id string) (entity.Workspace, error) {
	m, err := scanWorkspace(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+selectColumns+` FROM workspaces WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Workspace{}, entity.ErrWorkspaceNotFound
		}
		return entity.Workspace{}, fmt.Errorf("get workspace by id: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetBySlug(ctx context.Context, slug string) (entity.Workspace, error) {
	m, err := scanWorkspace(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+selectColumns+` FROM workspaces WHERE slug = $1`, slug))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Workspace{}, entity.ErrWorkspaceNotFound
		}
		return entity.Workspace{}, fmt.Errorf("get workspace by slug: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) Update(ctx context.Context, id string, u entity.WorkspaceUpdate) error {
	res, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE workspaces SET name = $2, timezone = $3, environment = $4, updated_at = now() WHERE id = $1`,
		id, u.Name, u.Timezone, u.Environment)
	if err != nil {
		return fmt.Errorf("update workspace: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update workspace rows: %w", err)
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
