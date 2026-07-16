package internal

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"

	"github.com/opsybot/opsybot/db"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
)

type Migrator struct {
	Log *slog.Logger
	PG  postgres.Client
}

func (m *Migrator) Migrate(ctx context.Context) error {
	ctx = logger.Into(ctx, m.Log)

	fsys, err := db.PostgresMigrations()
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}

	p, err := goose.NewProvider(goose.DialectPostgres, m.PG.DB, fsys)
	if err != nil {
		return fmt.Errorf("new migration provider: %w", err)
	}

	results, err := p.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	for _, r := range results {
		m.Log.InfoContext(ctx, "migration applied", "version", r.Source.Version, "name", r.Source.Path, "duration", r.Duration)
	}
	m.Log.InfoContext(ctx, "migrations up to date", "applied", len(results))
	return nil
}
