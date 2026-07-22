package alert_source

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

func New(db postgres.Client) repository.AlertSource {
	return &repo{db: db}
}

func toEntity(m *dbpostgres.AlertSource, mappings []entity.SourceMapping) entity.AlertSource {
	return entity.AlertSource{
		ID:                    m.ID,
		WorkspaceID:           m.WorkspaceID,
		Slug:                  m.Slug,
		Name:                  m.Name,
		Format:                entity.SourceFormat(m.Format),
		IngestToken:           m.IngestToken,
		SigningSecret:         m.SigningSecret,
		SigningSecretPrevious: m.SigningSecretPrevious,
		SecretRotatedAt:       m.SecretRotatedAt.Time,
		RequireSignature:      m.RequireSignature,
		DefaultSeverity:       entity.AlertSeverity(m.DefaultSeverity),
		AutoResolveAfter:      time.Duration(m.AutoResolveAfterSeconds) * time.Second,
		Mapping:               mappings,
		LastEventAt:           m.LastEventAt.Time,
		FailureCount:          m.FailureCount,
		Paused:                m.PausedAt.Valid,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}

func mappingsToEntity(rows dbpostgres.AlertSourceMappingSlice) []entity.SourceMapping {
	out := make([]entity.SourceMapping, 0, len(rows))
	for _, r := range rows {
		out = append(out, entity.SourceMapping{Field: r.Field, Path: r.Path, Position: r.Position})
	}
	return out
}

func (r *repo) loadMappings(ctx context.Context, sourceIDs ...string) (map[string][]entity.SourceMapping, error) {
	if len(sourceIDs) == 0 {
		return map[string][]entity.SourceMapping{}, nil
	}
	ids := make([]any, len(sourceIDs))
	for i, id := range sourceIDs {
		ids[i] = id
	}
	rows, err := dbpostgres.AlertSourceMappings(
		qm.WhereIn("source_id IN ?", ids...),
		qm.OrderBy("position, field"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list alert source mappings: %w", err)
	}
	out := make(map[string][]entity.SourceMapping, len(sourceIDs))
	for _, row := range rows {
		out[row.SourceID] = append(out[row.SourceID], entity.SourceMapping{Field: row.Field, Path: row.Path, Position: row.Position})
	}
	return out, nil
}

func (r *repo) Create(ctx context.Context, workspaceID string, src entity.AlertSource) (entity.AlertSource, error) {
	m := &dbpostgres.AlertSource{
		WorkspaceID:             workspaceID,
		Slug:                    src.Slug,
		Name:                    src.Name,
		Format:                  string(src.Format),
		IngestToken:             src.IngestToken,
		SigningSecret:           src.SigningSecret,
		RequireSignature:        src.RequireSignature,
		DefaultSeverity:         string(src.DefaultSeverity),
		AutoResolveAfterSeconds: int(src.AutoResolveAfter / time.Second),
	}
	cols := boil.Whitelist("workspace_id", "slug", "name", "format", "ingest_token", "signing_secret",
		"require_signature", "default_severity", "auto_resolve_after_seconds")
	if err := m.Insert(ctx, r.db.Querier(ctx), cols); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.AlertSource{}, entity.ErrAlertSourceSlugTaken
		}
		return entity.AlertSource{}, fmt.Errorf("create alert source: %w", err)
	}
	return toEntity(m, nil), nil
}

func (r *repo) Update(ctx context.Context, workspaceID, slug string, in entity.AlertSourceUpdate) (entity.AlertSource, error) {
	m, err := r.find(ctx, workspaceID, slug)
	if err != nil {
		return entity.AlertSource{}, err
	}
	m.Name = in.Name
	m.DefaultSeverity = string(in.DefaultSeverity)
	m.RequireSignature = in.RequireSignature
	m.AutoResolveAfterSeconds = int(in.AutoResolveAfter / time.Second)
	m.UpdatedAt = time.Now().UTC()
	cols := boil.Whitelist("name", "default_severity", "require_signature", "auto_resolve_after_seconds", "updated_at")
	if _, err := m.Update(ctx, r.db.Querier(ctx), cols); err != nil {
		return entity.AlertSource{}, fmt.Errorf("update alert source: %w", err)
	}
	return toEntity(m, nil), nil
}

func (r *repo) Delete(ctx context.Context, workspaceID, slug string) error {
	m, err := r.find(ctx, workspaceID, slug)
	if err != nil {
		return err
	}
	if _, err := m.Delete(ctx, r.db.Querier(ctx)); err != nil {
		return fmt.Errorf("delete alert source: %w", err)
	}
	return nil
}

func (r *repo) find(ctx context.Context, workspaceID, slug string) (*dbpostgres.AlertSource, error) {
	m, err := dbpostgres.AlertSources(
		qm.Where("workspace_id = ? AND slug = ?", workspaceID, slug),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrAlertSourceNotFound
		}
		return nil, fmt.Errorf("get alert source: %w", err)
	}
	return m, nil
}

func (r *repo) GetBySlug(ctx context.Context, workspaceID, slug string) (entity.AlertSource, error) {
	m, err := r.find(ctx, workspaceID, slug)
	if err != nil {
		return entity.AlertSource{}, err
	}
	mappings, err := r.loadMappings(ctx, m.ID)
	if err != nil {
		return entity.AlertSource{}, err
	}
	return toEntity(m, mappings[m.ID]), nil
}

func (r *repo) GetByToken(ctx context.Context, token string) (entity.AlertSource, error) {
	m, err := dbpostgres.AlertSources(qm.Where("ingest_token = ?", token)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AlertSource{}, entity.ErrAlertSourceNotFound
		}
		return entity.AlertSource{}, fmt.Errorf("get alert source by token: %w", err)
	}
	mappings, err := r.loadMappings(ctx, m.ID)
	if err != nil {
		return entity.AlertSource{}, err
	}
	return toEntity(m, mappings[m.ID]), nil
}

func (r *repo) ListByWorkspace(ctx context.Context, workspaceID string) ([]entity.AlertSource, error) {
	rows, err := dbpostgres.AlertSources(
		qm.Where("workspace_id = ?", workspaceID),
		qm.OrderBy("slug"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list alert sources: %w", err)
	}
	ids := make([]string, 0, len(rows))
	for _, m := range rows {
		ids = append(ids, m.ID)
	}
	mappings, err := r.loadMappings(ctx, ids...)
	if err != nil {
		return nil, err
	}
	out := make([]entity.AlertSource, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEntity(m, mappings[m.ID]))
	}
	return out, nil
}

func (r *repo) SetPaused(ctx context.Context, workspaceID, slug string, paused bool) error {
	m, err := r.find(ctx, workspaceID, slug)
	if err != nil {
		return err
	}
	if paused {
		m.PausedAt = null.TimeFrom(time.Now().UTC())
	} else {
		m.PausedAt = null.NewTime(time.Time{}, false)
	}
	m.UpdatedAt = time.Now().UTC()
	if _, err := m.Update(ctx, r.db.Querier(ctx), boil.Whitelist("paused_at", "updated_at")); err != nil {
		return fmt.Errorf("set alert source paused: %w", err)
	}
	return nil
}

func (r *repo) RotateSecret(ctx context.Context, workspaceID, slug, secret string) (entity.AlertSource, error) {
	m, err := r.find(ctx, workspaceID, slug)
	if err != nil {
		return entity.AlertSource{}, err
	}
	now := time.Now().UTC()
	m.SigningSecretPrevious = m.SigningSecret
	m.SigningSecret = secret
	m.SecretRotatedAt = null.TimeFrom(now)
	m.UpdatedAt = now
	cols := boil.Whitelist("signing_secret", "signing_secret_previous", "secret_rotated_at", "updated_at")
	if _, err := m.Update(ctx, r.db.Querier(ctx), cols); err != nil {
		return entity.AlertSource{}, fmt.Errorf("rotate alert source secret: %w", err)
	}
	return toEntity(m, nil), nil
}

func (r *repo) ReplaceMappings(ctx context.Context, sourceID string, mappings []entity.SourceMapping) error {
	exec := r.db.Querier(ctx)
	if _, err := dbpostgres.AlertSourceMappings(qm.Where("source_id = ?", sourceID)).DeleteAll(ctx, exec); err != nil {
		return fmt.Errorf("clear alert source mappings: %w", err)
	}
	for i, mapping := range mappings {
		row := &dbpostgres.AlertSourceMapping{
			SourceID: sourceID,
			Field:    mapping.Field,
			Path:     mapping.Path,
			Position: i,
		}
		if err := row.Insert(ctx, exec, boil.Whitelist("source_id", "field", "path", "position")); err != nil {
			return fmt.Errorf("insert alert source mapping: %w", err)
		}
	}
	return nil
}

func (r *repo) MarkDelivery(ctx context.Context, sourceID string, at time.Time, failed bool) error {
	values := dbpostgres.M{"last_event_at": at, "updated_at": at}
	if failed {
		if _, err := r.db.Querier(ctx).ExecContext(ctx,
			`UPDATE alert_sources SET last_event_at = $2, failure_count = failure_count + 1, updated_at = $2 WHERE id = $1`,
			sourceID, at); err != nil {
			return fmt.Errorf("mark alert source failure: %w", err)
		}
		return nil
	}
	values["failure_count"] = 0
	if _, err := dbpostgres.AlertSources(qm.Where("id = ?", sourceID)).UpdateAll(ctx, r.db.Querier(ctx), values); err != nil {
		return fmt.Errorf("mark alert source delivery: %w", err)
	}
	return nil
}
