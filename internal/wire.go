//go:build wireinject

package internal

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/pkg"
	"github.com/opsybot/opsybot/internal/repository"
)

var baseSet = wire.NewSet(
	config.Set,
	pkg.Set,
	repository.Set,
	wire.Struct(new(App), "*"),
	wire.Struct(new(Migrator), "*"),
)

func InitApp(cfgFile string) (*App, func(), error) {
	wire.Build(baseSet)
	return nil, nil, nil
}

func InitMigrator(cfgFile string) (*Migrator, func(), error) {
	wire.Build(baseSet)
	return nil, nil, nil
}
