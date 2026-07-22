package cron

import (
	"fmt"
	"log/slog"
	"time"

	redislock "github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/go-redsync/redsync/v4"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/pkg/valkey"
)

const (
	lockKeyPrefix     = "opsybot:cron:"
	lockExtendDivisor = 3
	defaultLockExpiry = time.Minute
)

type Client struct{ gocron.Scheduler }

func New(cfg config.Cron, vk valkey.Client, log *slog.Logger) (Client, func(), error) {
	expiry := cfg.LockExpiry
	if expiry <= 0 {
		expiry = defaultLockExpiry
	}

	locker, err := redislock.NewRedisLockerWithOptions(vk.UniversalClient,
		redislock.WithKeyPrefix(lockKeyPrefix),
		redislock.WithAutoExtendDuration(expiry/lockExtendDivisor),
		redislock.WithRedsyncOptions(redsync.WithExpiry(expiry), redsync.WithTries(1)),
	)
	if err != nil {
		return Client{}, nil, fmt.Errorf("build cron locker: %w", err)
	}

	scheduler, err := gocron.NewScheduler(
		gocron.WithDistributedLocker(locker),
		gocron.WithLocation(time.UTC),
		gocron.WithLogger(schedulerLogger{log: log}),
		gocron.WithStopTimeout(cfg.StopTimeout),
		gocron.WithGlobalJobOptions(gocron.WithSingletonMode(gocron.LimitModeReschedule)),
	)
	if err != nil {
		return Client{}, nil, fmt.Errorf("build cron scheduler: %w", err)
	}

	return Client{scheduler}, func() { _ = scheduler.Shutdown() }, nil
}
