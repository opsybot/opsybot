package internal

import (
	"context"
	"log/slog"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/cron"
	"github.com/opsybot/opsybot/internal/pkg/casbin"
	pkgcron "github.com/opsybot/opsybot/internal/pkg/cron"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/pkg/otel"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
)

type Worker struct {
	OTel        otel.Client
	Cfg         config.Config
	Log         *slog.Logger
	PG          postgres.Client
	Enforcer    casbin.Client
	Heartbeats  *cron.HeartbeatSweep
	AutoResolve *cron.AlertAutoResolve
	Retention   *cron.IngestRetention
}

func (w *Worker) Run(ctx context.Context) error {
	ctx = logger.Into(ctx, w.Log)

	jobs := pkgcron.New().With(
		pkgcron.Job{
			Name:     "heartbeat_sweep",
			Every:    w.Cfg.Cron.HeartbeatSweep,
			Deadline: w.Cfg.Cron.JobTimeout,
			RunOnce:  true,
			Run:      w.Heartbeats.Run,
		},
		pkgcron.Job{
			Name:     "alert_autoresolve",
			Every:    w.Cfg.Cron.AlertAutoResolve,
			Deadline: w.Cfg.Cron.JobTimeout,
			Run:      w.AutoResolve.Run,
		},
		pkgcron.Job{
			Name:     "ingest_retention",
			Every:    w.Cfg.Cron.IngestRetention,
			Deadline: w.Cfg.Cron.JobTimeout,
			Run:      w.Retention.Run,
		},
	)

	w.Log.InfoContext(ctx, "opsybot worker starting", "environment", w.Cfg.Environment)
	jobs.Run(ctx)
	w.Log.InfoContext(ctx, "worker stopped")
	return nil
}
