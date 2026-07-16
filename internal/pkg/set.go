package pkg

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/pkg/casbin"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/valkey"
)

var Set = wire.NewSet(
	logger.New,
	postgres.New,
	valkey.New,
	casbin.New,
)
