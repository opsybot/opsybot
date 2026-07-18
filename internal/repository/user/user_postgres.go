package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
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

func New(db postgres.Client, box secretbox.Client) repository.User {
	return &repo{db: db, box: box}
}

func toEntity(m *dbpostgres.User) entity.User {
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
	m := &dbpostgres.User{
		Email:        entity.NormalizeEmail(u.Email),
		Name:         u.Name,
		PasswordHash: null.StringFrom(passwordHash),
		Timezone:     u.Timezone,
	}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("email", "name", "password_hash", "timezone")); err != nil {
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
	m := &dbpostgres.User{Email: entity.NormalizeEmail(email), Name: name}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("email", "name")); err != nil {
		if n, ok := postgres.UniqueViolation(err); ok && n == "users_email_lower_uq" {
			return entity.User{}, entity.ErrUserEmailTaken
		}
		return entity.User{}, fmt.Errorf("create invited user: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) CreateSSO(ctx context.Context, email, name string) (entity.User, error) {
	m := &dbpostgres.User{Email: entity.NormalizeEmail(email), Name: name}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("email", "name")); err != nil {
		if n, ok := postgres.UniqueViolation(err); ok && n == "users_email_lower_uq" {
			return entity.User{}, entity.ErrUserEmailTaken
		}
		return entity.User{}, fmt.Errorf("create sso user: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetByID(ctx context.Context, id string) (entity.User, error) {
	m, err := dbpostgres.FindUser(ctx, r.db.Querier(ctx), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}
		return entity.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	m, err := dbpostgres.Users(qm.Where("lower(email) = ?", entity.NormalizeEmail(email))).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}
		return entity.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return toEntity(m), nil
}

func (r *repo) Activate(ctx context.Context, id, name, passwordHash, timezone string) error {
	if _, err := dbpostgres.Users(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"name": name, "password_hash": passwordHash, "timezone": timezone, "updated_at": time.Now()}); err != nil {
		return fmt.Errorf("activate user: %w", err)
	}
	return nil
}

func (r *repo) HasPassword(ctx context.Context, id string) (bool, error) {
	m, err := dbpostgres.Users(qm.Select("password_hash"), qm.Where("id = ?", id)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, entity.ErrUserNotFound
		}
		return false, fmt.Errorf("user has password: %w", err)
	}
	return m.PasswordHash.Valid, nil
}

func (r *repo) PasswordHash(ctx context.Context, id string) (string, error) {
	m, err := dbpostgres.Users(qm.Select("password_hash"), qm.Where("id = ?", id)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.ErrUserNotFound
		}
		return "", fmt.Errorf("get password hash: %w", err)
	}
	if !m.PasswordHash.Valid {
		return "", entity.ErrUserNoPassword
	}
	return m.PasswordHash.String, nil
}

func (r *repo) ExistsAny(ctx context.Context) (bool, error) {
	exists, err := dbpostgres.Users().Exists(ctx, r.db.Querier(ctx))
	if err != nil {
		return false, fmt.Errorf("users exist: %w", err)
	}
	return exists, nil
}

func (r *repo) TouchLastActive(ctx context.Context, id string) error {
	if _, err := dbpostgres.Users(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"last_active_at": time.Now()}); err != nil {
		return fmt.Errorf("touch last active: %w", err)
	}
	return nil
}

func (r *repo) UpdateProfile(ctx context.Context, id string, p entity.ProfileUpdate) error {
	if _, err := dbpostgres.Users(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"name": p.Name, "timezone": p.Timezone, "updated_at": time.Now()}); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

func (r *repo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	if _, err := dbpostgres.Users(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"password_hash": passwordHash, "updated_at": time.Now()}); err != nil {
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
	if _, err := dbpostgres.Users(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"totp_secret_enc": enc, "totp_enabled_at": nil, "totp_last_step": nil, "updated_at": time.Now()}); err != nil {
		return fmt.Errorf("set totp: %w", err)
	}
	return nil
}

func (r *repo) EnableTOTP(ctx context.Context, id string) error {
	now := time.Now()
	if _, err := dbpostgres.Users(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"totp_enabled_at": now, "updated_at": now}); err != nil {
		return fmt.Errorf("enable totp: %w", err)
	}
	return nil
}

func (r *repo) DisableTOTP(ctx context.Context, id string) error {
	if _, err := dbpostgres.Users(qm.Where("id = ?", id)).UpdateAll(ctx, r.db.Querier(ctx),
		dbpostgres.M{"totp_secret_enc": nil, "totp_enabled_at": nil, "totp_last_step": nil, "updated_at": time.Now()}); err != nil {
		return fmt.Errorf("disable totp: %w", err)
	}
	return nil
}

func (r *repo) TOTPSecret(ctx context.Context, id string) (string, error) {
	m, err := dbpostgres.Users(qm.Select("totp_secret_enc"), qm.Where("id = ?", id)).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.ErrUserNotFound
		}
		return "", fmt.Errorf("get totp secret: %w", err)
	}
	if !m.TotpSecretEnc.Valid || len(m.TotpSecretEnc.Bytes) == 0 {
		return "", entity.ErrTOTPNotEnrolled
	}
	plain, err := r.box.Decrypt(m.TotpSecretEnc.Bytes)
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
