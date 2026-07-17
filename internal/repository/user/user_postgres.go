package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	dbpostgres "github.com/opsybot/opsybot/internal/db/postgres"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const selectColumns = `id, email, name, password_hash, timezone, totp_enabled_at, last_active_at, created_at`

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.User {
	return &repo{db: db}
}

func scanUser(row interface {
	Scan(dest ...any) error
}) (dbpostgres.User, error) {
	var m dbpostgres.User
	err := row.Scan(&m.ID, &m.Email, &m.Name, &m.PasswordHash, &m.Timezone, &m.TotpEnabledAt, &m.LastActiveAt, &m.CreatedAt)
	return m, err
}

func toEntity(m dbpostgres.User) entity.User {
	return entity.User{
		ID:           m.ID,
		Email:        m.Email,
		Name:         m.Name,
		Timezone:     m.Timezone,
		HasPassword:  m.PasswordHash.Valid,
		TOTPEnabled:  m.TotpEnabledAt.Valid,
		LastActiveAt: m.LastActiveAt.Time,
		CreatedAt:    m.CreatedAt,
	}
}

func (r *repo) Create(ctx context.Context, u entity.NewUser, passwordHash string) (entity.User, error) {
	row := r.db.Querier(ctx).QueryRowContext(ctx,
		`INSERT INTO users (email, name, password_hash, timezone) VALUES ($1, $2, $3, $4) RETURNING `+selectColumns,
		entity.NormalizeEmail(u.Email), u.Name, passwordHash, u.Timezone)
	m, err := scanUser(row)
	if err != nil {
		if name, ok := postgres.UniqueViolation(err); ok && name == "users_email_lower_uq" {
			return entity.User{}, entity.ErrUserEmailTaken
		}
		return entity.User{}, fmt.Errorf("create user: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) CreateInvited(ctx context.Context, email string) (entity.User, error) {
	name := email
	if at := strings.IndexByte(email, '@'); at > 0 {
		name = email[:at]
	}
	row := r.db.Querier(ctx).QueryRowContext(ctx,
		`INSERT INTO users (email, name) VALUES ($1, $2) RETURNING `+selectColumns,
		entity.NormalizeEmail(email), name)
	m, err := scanUser(row)
	if err != nil {
		if n, ok := postgres.UniqueViolation(err); ok && n == "users_email_lower_uq" {
			return entity.User{}, entity.ErrUserEmailTaken
		}
		return entity.User{}, fmt.Errorf("create invited user: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetByID(ctx context.Context, id string) (entity.User, error) {
	m, err := scanUser(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+selectColumns+` FROM users WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}
		return entity.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	m, err := scanUser(r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT `+selectColumns+` FROM users WHERE lower(email) = $1`, entity.NormalizeEmail(email)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}
		return entity.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) PasswordHash(ctx context.Context, id string) (string, error) {
	var hash sql.NullString
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT password_hash FROM users WHERE id = $1`, id).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.ErrUserNotFound
		}
		return "", fmt.Errorf("get password hash: %w", err)
	}
	if !hash.Valid {
		return "", entity.ErrUserNoPassword
	}
	return hash.String, nil
}

func (r *repo) ExistsAny(ctx context.Context) (bool, error) {
	var exists bool
	err := r.db.Querier(ctx).QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users)`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("users exist: %w", err)
	}
	return exists, nil
}

func (r *repo) TouchLastActive(ctx context.Context, id string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE users SET last_active_at = now() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("touch last active: %w", err)
	}
	return nil
}
