package channel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/secretbox"
	"github.com/opsybot/opsybot/internal/repository"
)

type repo struct {
	db  postgres.Client
	box secretbox.Client
}

func New(db postgres.Client, box secretbox.Client) repository.Channel {
	return &repo{db: db, box: box}
}

func toEntity(m *dbpostgres.UserChannel) entity.Channel {
	return entity.Channel{
		ID:        m.ID,
		UserID:    m.UserID,
		Type:      entity.ChannelType(m.Type),
		Detail:    m.Detail,
		Label:     m.Label,
		Verified:  m.VerifiedAt.Valid,
		CreatedAt: m.CreatedAt,
	}
}

func (r *repo) Create(ctx context.Context, userID string, c entity.NewChannel) (entity.Channel, error) {
	m := &dbpostgres.UserChannel{UserID: userID, Type: string(c.Type), Detail: c.Detail, Label: c.Label}
	cols := []string{"user_id", "type", "detail", "label"}
	if c.Secret != "" {
		if !r.box.Enabled() {
			return entity.Channel{}, entity.ErrChannelSecretUnavailable
		}
		sealed, err := r.box.Encrypt([]byte(c.Secret))
		if err != nil {
			return entity.Channel{}, fmt.Errorf("seal channel secret: %w", err)
		}
		m.SecretEnc.Bytes = sealed
		m.SecretEnc.Valid = true
		cols = append(cols, "secret_enc")
	}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist(cols...)); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.Channel{}, entity.ErrChannelDuplicate
		}
		return entity.Channel{}, fmt.Errorf("create channel: %w", err)
	}
	out := toEntity(m)
	out.Secret = c.Secret
	return out, nil
}

func (r *repo) Secret(ctx context.Context, id, userID string) (string, error) {
	m, err := dbpostgres.UserChannels(
		qm.Select("secret_enc"),
		qm.Where("id = ? AND user_id = ?", id, userID),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.ErrChannelNotFound
		}
		return "", fmt.Errorf("get channel secret: %w", err)
	}
	if len(m.SecretEnc.Bytes) == 0 {
		return "", nil
	}
	if !r.box.Enabled() {
		return "", entity.ErrChannelSecretUnavailable
	}
	plain, err := r.box.Decrypt(m.SecretEnc.Bytes)
	if err != nil {
		return "", fmt.Errorf("decrypt channel secret: %w", err)
	}
	return string(plain), nil
}

func (r *repo) ListByUser(ctx context.Context, userID string) ([]entity.Channel, error) {
	rows, err := dbpostgres.UserChannels(
		qm.Where("user_id = ?", userID),
		qm.OrderBy("created_at"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	out := make([]entity.Channel, 0, len(rows))
	for _, m := range rows {
		out = append(out, toEntity(m))
	}
	return out, nil
}

func (r *repo) ListByUsers(ctx context.Context, userIDs []string) (map[string][]entity.Channel, error) {
	out := make(map[string][]entity.Channel, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	ids := make([]any, len(userIDs))
	for i, id := range userIDs {
		ids[i] = id
	}
	rows, err := dbpostgres.UserChannels(
		qm.WhereIn("user_id IN ?", ids...),
		qm.OrderBy("user_id, created_at, id"),
	).All(ctx, r.db.Querier(ctx))
	if err != nil {
		return nil, fmt.Errorf("list channels by users: %w", err)
	}
	for _, m := range rows {
		out[m.UserID] = append(out[m.UserID], toEntity(m))
	}
	return out, nil
}

func (r *repo) Get(ctx context.Context, id, userID string) (entity.Channel, error) {
	m, err := dbpostgres.UserChannels(qm.Where("id = ? AND user_id = ?", id, userID)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Channel{}, entity.ErrChannelNotFound
		}
		return entity.Channel{}, fmt.Errorf("get channel: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) MarkVerified(ctx context.Context, id, userID string) error {
	now := time.Now()
	n, err := dbpostgres.UserChannels(qm.Where("id = ? AND user_id = ?", id, userID)).
		UpdateAll(ctx, r.db.Querier(ctx), dbpostgres.M{"verified_at": now, "updated_at": now})
	if err != nil {
		return fmt.Errorf("verify channel: %w", err)
	}
	if n == 0 {
		return entity.ErrChannelNotFound
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, id, userID string) error {
	n, err := dbpostgres.UserChannels(qm.Where("id = ? AND user_id = ?", id, userID)).DeleteAll(ctx, r.db.Querier(ctx))
	if err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	if n == 0 {
		return entity.ErrChannelNotFound
	}
	return nil
}
