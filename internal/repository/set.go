package repository

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/pkg/postgres"
)

var Set = wire.NewSet(
	wire.Bind(new(Transactor), new(postgres.Client)),
)
