//go:build wireinject

package internal

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/handler"
	"github.com/opsybot/opsybot/internal/pkg"
)

var baseSet = wire.NewSet(
	config.Set,
	pkg.Set,
	repositoryProviders,
	serviceProviders,
	cronProviders,
	handler.Set,
	wire.Struct(new(App), "*"),
	wire.Struct(new(Migrator), "*"),
	wire.Struct(new(Worker), "*"),
)

func InitApp(cfgFile string) (*App, func(), error) {
	wire.Build(baseSet)
	return nil, nil, nil
}

func InitMigrator(cfgFile string) (*Migrator, func(), error) {
	wire.Build(baseSet)
	return nil, nil, nil
}

func InitWorker(cfgFile string) (*Worker, func(), error) {
	wire.Build(baseSet)
	return nil, nil, nil
}
