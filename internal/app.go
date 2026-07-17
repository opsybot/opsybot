package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/pkg/casbin"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/pkg/otel"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
)

type App struct {
	OTel     otel.Client
	Cfg      config.Config
	Log      *slog.Logger
	PG       postgres.Client
	Enforcer casbin.Client
	Router   http.Handler
}

func (a *App) Run(ctx context.Context) error {
	ctx = logger.Into(ctx, a.Log)

	addr := net.JoinHostPort(a.Cfg.HTTP.Host, strconv.Itoa(a.Cfg.HTTP.Port))
	srv := &http.Server{
		Addr:              addr,
		Handler:           a.Router,
		ReadHeaderTimeout: a.Cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       a.Cfg.HTTP.ReadTimeout,
		WriteTimeout:      a.Cfg.HTTP.WriteTimeout,
		IdleTimeout:       a.Cfg.HTTP.IdleTimeout,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	serveErr := make(chan error, 1)
	go func() {
		a.Log.InfoContext(ctx, "opsybot serving", "environment", a.Cfg.Environment, "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("serve http: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	a.Log.InfoContext(ctx, "shutting down", "timeout", a.Cfg.ShutdownTimeout)
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.Cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http: %w", err)
	}
	return <-serveErr
}
