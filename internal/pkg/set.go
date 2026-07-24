package pkg

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/pkg/casbin"
	"github.com/opsybot/opsybot/internal/pkg/cron"
	"github.com/opsybot/opsybot/internal/pkg/discord"
	"github.com/opsybot/opsybot/internal/pkg/logger"
	"github.com/opsybot/opsybot/internal/pkg/mailer"
	"github.com/opsybot/opsybot/internal/pkg/ntfy"
	"github.com/opsybot/opsybot/internal/pkg/objectstore"
	"github.com/opsybot/opsybot/internal/pkg/otel"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/secretbox"
	"github.com/opsybot/opsybot/internal/pkg/slack"
	pkgteams "github.com/opsybot/opsybot/internal/pkg/teams"
	"github.com/opsybot/opsybot/internal/pkg/telegram"
	"github.com/opsybot/opsybot/internal/pkg/valkey"
	"github.com/opsybot/opsybot/internal/pkg/webhook"
)

var Set = wire.NewSet(
	otel.New,
	logger.New,
	postgres.New,
	valkey.New,
	casbin.New,
	cron.New,
	mailer.New,
	secretbox.New,
	objectstore.New,
	webhook.New,
	ntfy.New,
	slack.New,
	discord.New,
	telegram.New,
	pkgteams.New,
)
