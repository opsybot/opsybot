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
	OTel          otel.Client
	Cfg           config.Config
	Log           *slog.Logger
	PG            postgres.Client
	Enforcer      casbin.Client
	Scheduler     pkgcron.Client
	Heartbeats    *cron.HeartbeatSweep
	AutoResolve   *cron.AlertAutoResolve
	Retention     *cron.IngestRetention
	Escalations   *cron.EscalationSweep
	Notifications *cron.NotificationSweep
}

func (w *Worker) Run(ctx context.Context) error {
	ctx = logger.Into(ctx, w.Log)

	jobs := []pkgcron.Job{
		{
			Name:    "heartbeat_sweep",
			Every:   w.Cfg.Cron.HeartbeatSweep,
			Timeout: w.Cfg.Cron.JobTimeout,
			AtStart: true,
			Run:     w.Heartbeats.Run,
		},
		{
			Name:    "alert_autoresolve",
			Every:   w.Cfg.Cron.AlertAutoResolve,
			Timeout: w.Cfg.Cron.JobTimeout,
			Run:     w.AutoResolve.Run,
		},
		{
			Name:    "ingest_retention",
			Crontab: w.Cfg.Cron.IngestRetention,
			Timeout: w.Cfg.Cron.JobTimeout,
			Run:     w.Retention.Run,
		},
		{
			Name:    "escalation_sweep",
			Every:   w.Cfg.Cron.EscalationSweep,
			Timeout: w.Cfg.Cron.JobTimeout,
			AtStart: true,
			Run:     w.Escalations.Run,
		},
		{
			Name:    "notification_sweep",
			Every:   w.Cfg.Cron.NotificationSweep,
			Timeout: w.Cfg.Cron.JobTimeout,
			AtStart: true,
			Run:     w.Notifications.Run,
		},
	}

	for _, job := range jobs {
		if _, err := w.Scheduler.Add(ctx, w.Log, job); err != nil {
			return err
		}
	}

	w.Scheduler.Start()
	w.Log.InfoContext(ctx, "opsybot worker started", "environment", w.Cfg.Environment)
	for _, job := range w.Scheduler.Jobs() {
		next, err := job.NextRun()
		if err != nil {
			return err
		}
		w.Log.InfoContext(ctx, "cron job scheduled", "job", job.Name(), "next_run", next)
	}

	<-ctx.Done()
	w.Log.InfoContext(ctx, "worker stopping", "timeout", w.Cfg.Cron.StopTimeout)
	return nil
}
