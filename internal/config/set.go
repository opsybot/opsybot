package config

import "github.com/goforj/wire"

var Set = wire.NewSet(
	New,
	NewPostgres,
	NewLog,
	NewOTel,
	NewEnvironment,
	NewAuth,
	NewMailer,
	NewHTTP,
	NewValkey,
	NewCasbin,
	NewIngest,
	NewCron,
	NewWebhook,
	NewNtfy,
	NewSlack,
	NewDiscord,
	NewTelegram,
	NewTeams,
	NewChat,
)
