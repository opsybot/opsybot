package cron

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/pkg/logger"
)

type Job struct {
	Name     string
	Every    time.Duration
	Run      func(ctx context.Context, now time.Time) (int, error)
	RunOnce  bool
	Deadline time.Duration
}

type Client struct {
	jobs []Job
}

func New() Client {
	return Client{}
}

func (c Client) With(jobs ...Job) Client {
	return Client{jobs: append(append([]Job{}, c.jobs...), jobs...)}
}

func (c Client) Run(ctx context.Context) {
	done := make(chan struct{}, len(c.jobs))
	for _, job := range c.jobs {
		go func() {
			defer func() { done <- struct{}{} }()
			c.loop(ctx, job)
		}()
	}
	for range c.jobs {
		<-done
	}
}

func (c Client) loop(ctx context.Context, job Job) {
	ticker := time.NewTicker(job.Every)
	defer ticker.Stop()

	if job.RunOnce {
		c.tick(ctx, job)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.tick(ctx, job)
		}
	}
}

func (c Client) tick(ctx context.Context, job Job) {
	log := logger.From(ctx).With("job", job.Name)

	runCtx := ctx
	if job.Deadline > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, job.Deadline)
		defer cancel()
	}

	started := time.Now().UTC()
	affected, err := job.Run(runCtx, started)
	if err != nil {
		log.ErrorContext(ctx, "cron job failed", "error", err, "duration", time.Since(started))
		return
	}
	if affected > 0 {
		log.InfoContext(ctx, "cron job finished", "affected", affected, "duration", time.Since(started))
	}
}
