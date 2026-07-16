package config

import "github.com/goforj/wire"

var Set = wire.NewSet(
	New,
	NewPostgres,
	NewLog,
	NewValkey,
	NewCasbin,
)
