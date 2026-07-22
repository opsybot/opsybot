package silence

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

func New(db postgres.Client) repository.Silence {
	return &repo{db: db}
}

func (r *repo) load(ctx context.Context, mods ...qm.QueryMod) ([]entity.Silence, error) {
	exec := r.db.Querier(ctx)
	rows, err := dbpostgres.AlertSilences(append(mods, qm.OrderBy("starts_at DESC, id DESC"))...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list silences: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}

	ids := make([]any, 0, len(rows))
	for _, m := range rows {
		ids = append(ids, m.ID)
	}
	conds, err := dbpostgres.AlertSilenceConditions(
		qm.WhereIn("silence_id IN ?", ids...),
		qm.OrderBy("position, id"),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list silence conditions: %w", err)
	}
	bySilence := make(map[string][]entity.SilenceCondition, len(rows))
	for _, c := range conds {
		bySilence[c.SilenceID] = append(bySilence[c.SilenceID], entity.SilenceCondition{Field: c.Field, Value: c.Value})
	}

	out := make([]entity.Silence, 0, len(rows))
	for _, m := range rows {
		out = append(out, entity.Silence{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			Kind:        entity.SilenceKind(m.Kind),
			Reason:      m.Reason,
			CreatedBy:   m.CreatedBy,
			Conditions:  bySilence[m.ID],
			StartsAt:    m.StartsAt,
			EndsAt:      m.EndsAt,
			CreatedAt:   m.CreatedAt,
		})
	}
	return out, nil
}

func (r *repo) List(ctx context.Context, workspaceID string) ([]entity.Silence, error) {
	return r.load(ctx, qm.Where("workspace_id = ?", workspaceID))
}

func (r *repo) ListActive(ctx context.Context, workspaceID string, at time.Time) ([]entity.Silence, error) {
	return r.load(ctx, qm.Where("workspace_id = ? AND starts_at <= ? AND ends_at > ?", workspaceID, at, at))
}

func (r *repo) Create(ctx context.Context, workspaceID, createdBy, createdByUserID string, in entity.NewSilence) (entity.Silence, error) {
	exec := r.db.Querier(ctx)
	kind := in.Kind
	if kind == "" {
		kind = entity.SilenceKindSilence
	}
	m := &dbpostgres.AlertSilence{
		WorkspaceID: workspaceID,
		Kind:        string(kind),
		Reason:      in.Reason,
		CreatedBy:   createdBy,
		StartsAt:    in.StartsAt,
		EndsAt:      in.EndsAt,
	}
	if createdByUserID != "" {
		m.CreatedByUserID = null.StringFrom(createdByUserID)
	}
	cols := boil.Whitelist("workspace_id", "kind", "reason", "created_by", "created_by_user_id", "starts_at", "ends_at")
	if err := m.Insert(ctx, exec, cols); err != nil {
		return entity.Silence{}, fmt.Errorf("create silence: %w", err)
	}
	for i, c := range in.Conditions {
		row := &dbpostgres.AlertSilenceCondition{SilenceID: m.ID, Field: c.Field, Value: c.Value, Position: i}
		if err := row.Insert(ctx, exec, boil.Whitelist("silence_id", "field", "value", "position")); err != nil {
			return entity.Silence{}, fmt.Errorf("insert silence condition: %w", err)
		}
	}
	return entity.Silence{
		ID:          m.ID,
		WorkspaceID: workspaceID,
		Kind:        kind,
		Reason:      in.Reason,
		CreatedBy:   createdBy,
		Conditions:  in.Conditions,
		StartsAt:    in.StartsAt,
		EndsAt:      in.EndsAt,
		CreatedAt:   m.CreatedAt,
	}, nil
}

func (r *repo) End(ctx context.Context, workspaceID, silenceID string, at time.Time) error {
	m, err := dbpostgres.AlertSilences(qm.Where("workspace_id = ? AND id = ?", workspaceID, silenceID)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrSilenceNotFound
		}
		return fmt.Errorf("get silence: %w", err)
	}
	if !m.EndsAt.After(at) {
		return entity.ErrSilenceEnded
	}
	m.EndsAt = at
	if _, err := m.Update(ctx, r.db.Querier(ctx), boil.Whitelist("ends_at")); err != nil {
		return fmt.Errorf("end silence: %w", err)
	}
	return nil
}
