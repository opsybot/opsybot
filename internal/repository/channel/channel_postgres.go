package channel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const selectColumns = `id, user_id, type, detail, verified_at, created_at`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Channel {
	return &repo{db: db}
}

func scanChannel(row interface {
	Scan(dest ...any) error
}) (entity.Channel, error) {
	var (
		c          entity.Channel
		typ        string
		verifiedAt sql.NullTime
	)
	if err := row.Scan(&c.ID, &c.UserID, &typ, &c.Detail, &verifiedAt, &c.CreatedAt); err != nil {
		return entity.Channel{}, err
	}
	c.Type = entity.ChannelType(typ)
	c.Verified = verifiedAt.Valid
	return c, nil
}

func (r *repo) Create(ctx context.Context, userID string, c entity.NewChannel) (entity.Channel, error) {
	ch, err := scanChannel(r.db.Querier(ctx).QueryRowContext(ctx,
		`INSERT INTO user_channels (user_id, type, detail) VALUES ($1, $2, $3) RETURNING `+selectColumns,
		userID, string(c.Type), c.Detail))
	if err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.Channel{}, entity.ErrChannelDuplicate
		}
		return entity.Channel{}, fmt.Errorf("create channel: %w", err)
	}
	return ch, nil
}

func (r *repo) ListByUser(ctx context.Context, userID string) ([]entity.Channel, error) {
	rows, err := r.db.Querier(ctx).QueryContext(ctx,
		`SELECT `+selectColumns+` FROM user_channels WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []entity.Channel
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		out = append(out, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate channels: %w", err)
	}
	return out, nil
}

func (r *repo) Get(ctx context.Context, id, userID string) (entity.Channel, error) {
	ch, err := scanChannel(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+selectColumns+` FROM user_channels WHERE id = $1 AND user_id = $2`, id, userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Channel{}, entity.ErrChannelNotFound
		}
		return entity.Channel{}, fmt.Errorf("get channel: %w", err)
	}
	return ch, nil
}

func (r *repo) MarkVerified(ctx context.Context, id, userID string) error {
	res, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE user_channels SET verified_at = now(), updated_at = now() WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("verify channel: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("verify channel rows: %w", err)
	}
	if n == 0 {
		return entity.ErrChannelNotFound
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, id, userID string) error {
	res, err := r.db.Querier(ctx).ExecContext(ctx,
		`DELETE FROM user_channels WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete channel rows: %w", err)
	}
	if n == 0 {
		return entity.ErrChannelNotFound
	}
	return nil
}
