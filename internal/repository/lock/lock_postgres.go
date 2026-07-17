package lock

import (
	"context"
	"fmt"

	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const instanceKey = "opsybot:instance:setup"

type repo struct {
	db postgres.Client
}

func New(db postgres.Client) repository.Lock {
	return &repo{db: db}
}

func (r *repo) Workspace(ctx context.Context, workspaceID string) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "opsybot:ws:"+workspaceID); err != nil {
		return fmt.Errorf("lock workspace: %w", err)
	}
	return nil
}

func (r *repo) Instance(ctx context.Context) error {
	if _, err := r.db.Querier(ctx).ExecContext(ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, instanceKey); err != nil {
		return fmt.Errorf("lock instance: %w", err)
	}
	return nil
}
