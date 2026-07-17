package config

import "github.com/goforj/wire"

var Set = wire.NewSet(
	New,
	NewPostgres,
	NewLog,
	NewOTel,
	NewEnvironment,
	NewAuth,
	NewHTTP,
	NewValkey,
	NewCasbin,
)
