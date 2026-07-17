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
	"github.com/opsybot/opsybot/internal/pkg/secretbox"
	"github.com/opsybot/opsybot/internal/repository"
)

const selectColumns = `id, email, name, password_hash, timezone, totp_enabled_at, last_active_at, created_at`

type repo struct {
	db  postgres.Client
	box secretbox.Client
}

func New(db postgres.Client, box secretbox.Client) repository.User {
	return &repo{db: db, box: box}
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

func (r *repo) Activate(ctx context.Context, id, name, passwordHash, timezone string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE users SET name = $2, password_hash = $3, timezone = $4, updated_at = now() WHERE id = $1`,
		id, name, passwordHash, timezone); err != nil {
		return fmt.Errorf("activate user: %w", err)
	}
	return nil
}

func (r *repo) HasPassword(ctx context.Context, id string) (bool, error) {
	var hasPassword bool
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT password_hash IS NOT NULL FROM users WHERE id = $1`, id).Scan(&hasPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, entity.ErrUserNotFound
		}
		return false, fmt.Errorf("user has password: %w", err)
	}
	return hasPassword, nil
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

func (r *repo) UpdateProfile(ctx context.Context, id string, p entity.ProfileUpdate) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE users SET name = $2, timezone = $3, updated_at = now() WHERE id = $1`,
		id, p.Name, p.Timezone); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

func (r *repo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`,
		id, passwordHash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (r *repo) SetTOTP(ctx context.Context, id, secret string) error {
	enc, err := r.box.Encrypt([]byte(secret))
	if err != nil {
		if errors.Is(err, secretbox.ErrDisabled) {
			return entity.ErrTOTPUnavailable
		}
		return err
	}
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE users SET totp_secret_enc = $2, totp_enabled_at = NULL, totp_last_step = NULL, updated_at = now() WHERE id = $1`,
		id, enc); err != nil {
		return fmt.Errorf("set totp: %w", err)
	}
	return nil
}

func (r *repo) EnableTOTP(ctx context.Context, id string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE users SET totp_enabled_at = now(), updated_at = now() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("enable totp: %w", err)
	}
	return nil
}

func (r *repo) DisableTOTP(ctx context.Context, id string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE users SET totp_secret_enc = NULL, totp_enabled_at = NULL, totp_last_step = NULL, updated_at = now() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("disable totp: %w", err)
	}
	return nil
}

func (r *repo) TOTPSecret(ctx context.Context, id string) (string, error) {
	var enc []byte
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT totp_secret_enc FROM users WHERE id = $1`, id).Scan(&enc)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.ErrUserNotFound
		}
		return "", fmt.Errorf("get totp secret: %w", err)
	}
	if len(enc) == 0 {
		return "", entity.ErrTOTPNotEnrolled
	}
	plain, err := r.box.Decrypt(enc)
	if err != nil {
		if errors.Is(err, secretbox.ErrDisabled) {
			return "", entity.ErrTOTPUnavailable
		}
		return "", err
	}
	return string(plain), nil
}

func (r *repo) AcceptTOTPStep(ctx context.Context, id string, step int64) (bool, error) {
	res, err := r.db.Querier(ctx).ExecContext(ctx,
		`UPDATE users SET totp_last_step = $2 WHERE id = $1 AND (totp_last_step IS NULL OR totp_last_step < $2)`,
		id, step)
	if err != nil {
		return false, fmt.Errorf("accept totp step: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("accept totp step rows: %w", err)
	}
	return n > 0, nil
}
