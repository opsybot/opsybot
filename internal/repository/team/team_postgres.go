package team

import (
	"context"
	"database/sql"
	"errors"
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

func New(db postgres.Client) repository.Team {
	return &repo{db: db}
}

func toEntity(m *dbpostgres.Team) entity.Team {
	return entity.Team{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		Slug:        m.Slug,
		Name:        m.Name,
		Archived:    m.ArchivedAt.Valid,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (r *repo) Create(ctx context.Context, workspaceID, slug, name string, memberIDs []string) (entity.Team, error) {
	m := &dbpostgres.Team{WorkspaceID: workspaceID, Slug: slug, Name: name}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("workspace_id", "slug", "name")); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.Team{}, entity.ErrTeamSlugTaken
		}
		return entity.Team{}, fmt.Errorf("create team: %w", err)
	}
	if err := r.replaceMembers(ctx, m.ID, workspaceID, memberIDs); err != nil {
		return entity.Team{}, err
	}
	t := toEntity(m)
	t.MemberIDs = memberIDs
	return t, nil
}

func (r *repo) GetBySlug(ctx context.Context, workspaceID, slug string) (entity.Team, error) {
	m, err := dbpostgres.Teams(qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Team{}, entity.ErrTeamNotFound
		}
		return entity.Team{}, fmt.Errorf("get team: %w", err)
	}
	t := toEntity(m)
	t.MemberIDs, err = r.memberIDs(ctx, t.ID)
	if err != nil {
		return entity.Team{}, err
	}
	return t, nil
}

func (r *repo) ListByWorkspace(ctx context.Context, workspaceID string, includeArchived bool) ([]entity.Team, error) {
	mods := []qm.QueryMod{qm.Where("workspace_id = ?", workspaceID)}
	if !includeArchived {
		mods = append(mods, qm.Where("archived_at IS NULL"))
	}
	mods = append(mods, qm.OrderBy("name"))
	rows, err := dbpostgres.Teams(mods...).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}

	teams := make([]entity.Team, 0, len(rows))
	index := make(map[string]int)
	for _, m := range rows {
		t := toEntity(m)
		t.MemberIDs = []string{}
		index[t.ID] = len(teams)
		teams = append(teams, t)
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
	exists, err := dbpostgres.Teams(qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug)).Exists(ctx, r.db.Querier(ctx))
	if err != nil {
		return false, fmt.Errorf("team slug exists: %w", err)
	}
	return exists, nil
}

func (r *repo) replaceMembers(ctx context.Context, teamID, workspaceID string, memberIDs []string) error {
	if _, err := dbpostgres.TeamMembers(qm.Where("team_id = ?", teamID)).DeleteAll(ctx, r.db.Querier(ctx)); err != nil {
		return fmt.Errorf("clear team members: %w", err)
	}
	for _, userID := range memberIDs {
		tm := &dbpostgres.TeamMember{TeamID: teamID, WorkspaceID: workspaceID, UserID: userID}
		if err := tm.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("team_id", "workspace_id", "user_id")); err != nil {
			return fmt.Errorf("add team member: %w", err)
		}
	}
	return nil
}

func (r *repo) memberIDs(ctx context.Context, teamID string) ([]string, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT tm.user_id FROM team_members tm
		 JOIN workspace_members wm ON wm.workspace_id = tm.workspace_id AND wm.user_id = tm.user_id AND wm.status = 'active'
		 WHERE tm.team_id = $1 ORDER BY tm.created_at`, teamID)
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
		`SELECT tm.team_id, tm.user_id FROM team_members tm
		 JOIN workspace_members wm ON wm.workspace_id = tm.workspace_id AND wm.user_id = tm.user_id AND wm.status = 'active'
		 WHERE tm.workspace_id = $1 ORDER BY tm.created_at`, workspaceID)
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
