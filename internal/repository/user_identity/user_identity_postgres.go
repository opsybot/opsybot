package user_identity

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

func New(db postgres.Client) repository.UserIdentity {
	return &repo{db: db}
}

func (r *repo) GetBySubject(ctx context.Context, connectionID, subject string) (entity.UserIdentity, error) {
	m, err := dbpostgres.UserIdentities(
		qm.Where("connection_id = ? AND subject = ?", connectionID, subject),
	).One(ctx, r.db.Querier(ctx))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserIdentity{}, entity.ErrUserIdentityNotFound
		}
		return entity.UserIdentity{}, fmt.Errorf("get user identity: %w", err)
	}
	return entity.UserIdentity{
		ID:           m.ID,
		UserID:       m.UserID,
		ConnectionID: m.ConnectionID,
		Subject:      m.Subject,
		Email:        m.Email,
	}, nil
}

func (r *repo) Create(ctx context.Context, userID, connectionID, subject, email string) error {
	m := &dbpostgres.UserIdentity{UserID: userID, ConnectionID: connectionID, Subject: subject, Email: email}
	if err := m.Insert(ctx, r.db.Querier(ctx), boil.Whitelist("user_id", "connection_id", "subject", "email")); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.ErrUserIdentityExists
		}
		return fmt.Errorf("create user identity: %w", err)
	}
	return nil
}
