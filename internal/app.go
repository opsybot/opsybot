package internal

import (
	"context"
	"log/slog"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
)

type App struct {
	Cfg config.Config
	Log *slog.Logger
	PG  postgres.Client
}

func (a *App) Run(ctx context.Context) error {
	ctx = logger.Into(ctx, a.Log)
	a.Log.InfoContext(ctx, "opsybot serving", "environment", a.Cfg.Environment)
	<-ctx.Done()
	a.Log.Info("shutting down")
	return nil
}
