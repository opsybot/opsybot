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

func (r *repo) TryJob(ctx context.Context, name string) (bool, error) {
	var acquired bool
	if err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT pg_try_advisory_xact_lock(hashtextextended($1, 0))`, "opsybot:job:"+name).Scan(&acquired); err != nil {
		return false, fmt.Errorf("try job lock: %w", err)
	}
	return acquired, nil
}
