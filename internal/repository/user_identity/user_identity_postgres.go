package user_identity

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

func New(db postgres.Client) repository.UserIdentity {
	return &repo{db: db}
}

func (r *repo) GetBySubject(ctx context.Context, connectionID, subject string) (entity.UserIdentity, error) {
	var ui entity.UserIdentity
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT id, user_id, connection_id, subject, email FROM user_identities
		 WHERE connection_id = $1 AND subject = $2`, connectionID, subject).
		Scan(&ui.ID, &ui.UserID, &ui.ConnectionID, &ui.Subject, &ui.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserIdentity{}, entity.ErrUserIdentityNotFound
		}
		return entity.UserIdentity{}, fmt.Errorf("get user identity: %w", err)
	}
	return ui, nil
}

func (r *repo) Create(ctx context.Context, userID, connectionID, subject, email string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`INSERT INTO user_identities (user_id, connection_id, subject, email) VALUES ($1, $2, $3, $4)`,
		userID, connectionID, subject, email); err != nil {
		if _, ok := postgres.UniqueViolation(err); ok {
			return entity.ErrUserIdentityExists
		}
		return fmt.Errorf("create user identity: %w", err)
	}
	return nil
}
