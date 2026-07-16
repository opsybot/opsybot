package pkg

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
)

var Set = wire.NewSet(
	logger.New,
	postgres.New,
)
