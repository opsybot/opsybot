package team

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Team {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, workspaceID, slug, name string, memberIDs []string) (entity.Team, error) {
	t := entity.Team{WorkspaceID: workspaceID, Slug: slug, Name: name}
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`INSERT INTO teams (workspace_id, slug, name) VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		workspaceID, slug, name).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.Team{}, entity.ErrTeamSlugTaken
		}
		return entity.Team{}, fmt.Errorf("create team: %w", err)
	}
	if err := r.replaceMembers(ctx, t.ID, workspaceID, memberIDs); err != nil {
		return entity.Team{}, err
	}
	t.MemberIDs = memberIDs
	return t, nil
}

func (r *repo) GetBySlug(ctx context.Context, workspaceID, slug string) (entity.Team, error) {
	var (
		t          entity.Team
		archivedAt sql.NullTime
	)
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT id, workspace_id, slug, name, archived_at, created_at, updated_at
		 FROM teams WHERE workspace_id = $1 AND slug = $2`, workspaceID, slug).
		Scan(&t.ID, &t.WorkspaceID, &t.Slug, &t.Name, &archivedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Team{}, entity.ErrTeamNotFound
		}
		return entity.Team{}, fmt.Errorf("get team: %w", err)
	}
	t.Archived = archivedAt.Valid
	t.MemberIDs, err = r.memberIDs(ctx, t.ID)
	if err != nil {
		return entity.Team{}, err
	}
	return t, nil
}

func (r *repo) ListByWorkspace(ctx context.Context, workspaceID string, includeArchived bool) ([]entity.Team, error) {
	filter := ""
	if !includeArchived {
		filter = " AND archived_at IS NULL"
	}
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT id, workspace_id, slug, name, archived_at, created_at, updated_at
		 FROM teams WHERE workspace_id = $1`+filter+` ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	teams := make([]entity.Team, 0)
	index := make(map[string]int)
	for rows.Next() {
		var (
			t          entity.Team
			archivedAt sql.NullTime
		)
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &t.Slug, &t.Name, &archivedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		t.Archived = archivedAt.Valid
		t.MemberIDs = []string{}
		index[t.ID] = len(teams)
		teams = append(teams, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teams: %w", err)
	}
	if len(teams) == 0 {
		return teams, nil
	}
	if err := r.attachMembers(ctx, workspaceID, index, teams); err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *repo) Update(ctx context.Context, workspaceID, slug, name string, memberIDs []string) (entity.Team, error) {
	var (
		t          entity.Team
		archivedAt sql.NullTime
	)
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`UPDATE teams SET name = $3, updated_at = now()
		 WHERE workspace_id = $1 AND slug = $2
		 RETURNING id, workspace_id, slug, name, archived_at, created_at, updated_at`,
		workspaceID, slug, name).
		Scan(&t.ID, &t.WorkspaceID, &t.Slug, &t.Name, &archivedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Team{}, entity.ErrTeamNotFound
		}
		return entity.Team{}, fmt.Errorf("update team: %w", err)
	}
	t.Archived = archivedAt.Valid
	if err := r.replaceMembers(ctx, t.ID, workspaceID, memberIDs); err != nil {
		return entity.Team{}, err
	}
	t.MemberIDs = memberIDs
	return t, nil
}

func (r *repo) SetArchived(ctx context.Context, workspaceID, slug string, archived bool) (entity.Team, error) {
	value := "now()"
	if !archived {
		value = "NULL"
	}
	var (
		t          entity.Team
		archivedAt sql.NullTime
	)
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`UPDATE teams SET archived_at = `+value+`, updated_at = now()
		 WHERE workspace_id = $1 AND slug = $2
		 RETURNING id, workspace_id, slug, name, archived_at, created_at, updated_at`,
		workspaceID, slug).
		Scan(&t.ID, &t.WorkspaceID, &t.Slug, &t.Name, &archivedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Team{}, entity.ErrTeamNotFound
		}
		return entity.Team{}, fmt.Errorf("set team archived: %w", err)
	}
	t.Archived = archivedAt.Valid
	t.MemberIDs, err = r.memberIDs(ctx, t.ID)
	if err != nil {
		return entity.Team{}, err
	}
	return t, nil
}

func (r *repo) SlugExists(ctx context.Context, workspaceID, slug string) (bool, error) {
	var exists bool
	if err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM teams WHERE workspace_id = $1 AND slug = $2)`,
		workspaceID, slug).Scan(&exists); err != nil {
		return false, fmt.Errorf("team slug exists: %w", err)
	}
	return exists, nil
}

func (r *repo) replaceMembers(ctx context.Context, teamID, workspaceID string, memberIDs []string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`DELETE FROM team_members WHERE team_id = $1`, teamID); err != nil {
		return fmt.Errorf("clear team members: %w", err)
	}
	for _, userID := range memberIDs {
		if _, err := r.db.Querier(ctx).ExecContext(ctx,
			`INSERT INTO team_members (team_id, workspace_id, user_id) VALUES ($1, $2, $3)`,
			teamID, workspaceID, userID); err != nil {
			return fmt.Errorf("add team member: %w", err)
		}
	}
	return nil
}

func (r *repo) memberIDs(ctx context.Context, teamID string) ([]string, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT user_id FROM team_members WHERE team_id = $1 ORDER BY created_at`, teamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team members: %w", err)
	}
	return out, nil
}

func (r *repo) attachMembers(ctx context.Context, workspaceID string, index map[string]int, teams []entity.Team) error {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT team_id, user_id FROM team_members WHERE workspace_id = $1 ORDER BY created_at`, workspaceID)
	if err != nil {
		return fmt.Errorf("list team members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var teamID, userID string
		if err := rows.Scan(&teamID, &userID); err != nil {
			return fmt.Errorf("scan team member: %w", err)
		}
		if i, ok := index[teamID]; ok {
			teams[i].MemberIDs = append(teams[i].MemberIDs, userID)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate team members: %w", err)
	}
	return nil
}
