package casbin

import (
	"context"
	_ "embed"
	"fmt"

	sqladapter "github.com/Blank-Xu/sql-adapter"
	casbinv3 "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	rediswatcher "github.com/casbin/redis-watcher/v2"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/valkey"
)

const driverName = "pgx"

//go:embed model.conf
var modelConf string

type Client struct{ *casbinv3.SyncedEnforcer }

func New(cfg config.Casbin, pg postgres.Client, vk valkey.Client) (Client, func(), error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return Client{}, nil, fmt.Errorf("parse casbin model: %w", err)
	}

	a, err := sqladapter.NewAdapterWithContext(context.Background(), pg.DB, driverName, cfg.TableName)
	if err != nil {
		return Client{}, nil, fmt.Errorf("new casbin adapter: %w", err)
	}

	e, err := casbinv3.NewSyncedEnforcer(m, a)
	if err != nil {
		return Client{}, nil, fmt.Errorf("new casbin enforcer: %w", err)
	}

	w, err := rediswatcher.NewWatcher("", watcherOptions(e, cfg, vk))
	if err != nil {
		return Client{}, nil, fmt.Errorf("new casbin watcher: %w", err)
	}

	if err := e.SetWatcher(w); err != nil {
		w.Close()
		return Client{}, nil, fmt.Errorf("set casbin watcher: %w", err)
	}

	return Client{e}, func() { w.Close() }, nil
}

func watcherOptions(e *casbinv3.SyncedEnforcer, cfg config.Casbin, vk valkey.Client) rediswatcher.WatcherOptions {
	return rediswatcher.WatcherOptions{
		SubClient:              vk,
		PubClient:              vk,
		Channel:                cfg.Channel,
		IgnoreSelf:             true,
		OptionalUpdateCallback: func(string) { _ = e.LoadPolicy() },
	}
}
