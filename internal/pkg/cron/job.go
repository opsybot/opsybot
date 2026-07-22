package cron

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

const cronFieldsWithSeconds = 6

var ErrJobSchedule = errors.New("cron job needs an interval or a crontab")

type Job struct {
	Name    string
	Every   time.Duration
	Crontab string
	Timeout time.Duration
	AtStart bool
	Run     func(ctx context.Context) (int, error)
}

func (j Job) definition() (gocron.JobDefinition, error) {
	if crontab := strings.TrimSpace(j.Crontab); crontab != "" {
		return gocron.CronJob(crontab, len(strings.Fields(crontab)) == cronFieldsWithSeconds), nil
	}
	if j.Every <= 0 {
		return nil, fmt.Errorf("%w: %s", ErrJobSchedule, j.Name)
	}
	return gocron.DurationJob(j.Every), nil
}

func (c Client) Add(ctx context.Context, log *slog.Logger, j Job) (gocron.Job, error) {
	definition, err := j.definition()
	if err != nil {
		return nil, err
	}

	options := []gocron.JobOption{
		gocron.WithName(j.Name),
		gocron.WithContext(ctx),
		gocron.WithEventListeners(
			gocron.AfterJobRunsWithError(func(_ uuid.UUID, name string, err error) {
				log.ErrorContext(ctx, "cron job failed", "job", name, "error", err)
			}),
			gocron.AfterJobRunsWithPanic(func(_ uuid.UUID, name string, recovered any) {
				log.ErrorContext(ctx, "cron job panicked", "job", name, "panic", recovered)
			}),
			gocron.AfterLockError(func(_ uuid.UUID, name string, err error) {
				log.DebugContext(ctx, "cron job skipped", "job", name, "reason", err)
			}),
		),
	}
	if j.AtStart {
		options = append(options, gocron.WithStartAt(gocron.WithStartImmediately()))
	}

	job, err := c.NewJob(definition, gocron.NewTask(j.task(log)), options...)
	if err != nil {
		return nil, fmt.Errorf("schedule cron job %s: %w", j.Name, err)
	}
	return job, nil
}

func (j Job) task(log *slog.Logger) func(context.Context) error {
	return func(ctx context.Context) error {
		if j.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, j.Timeout)
			defer cancel()
		}

		started := time.Now()
		affected, err := j.Run(ctx)
		if err != nil {
			return err
		}
		if affected > 0 {
			log.InfoContext(ctx, "cron job finished",
				"job", j.Name, "affected", affected, "duration", time.Since(started))
		}
		return nil
	}
}
