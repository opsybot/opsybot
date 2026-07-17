package internal

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/repository/audit"
	"github.com/opsybot/opsybot/internal/repository/channel"
	"github.com/opsybot/opsybot/internal/repository/invite"
	"github.com/opsybot/opsybot/internal/repository/lock"
	"github.com/opsybot/opsybot/internal/repository/mailer"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/password_reset"
	"github.com/opsybot/opsybot/internal/repository/pending"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/recovery_code"
	"github.com/opsybot/opsybot/internal/repository/session"
	"github.com/opsybot/opsybot/internal/repository/team"
	"github.com/opsybot/opsybot/internal/repository/transactor"
	"github.com/opsybot/opsybot/internal/repository/user"
	"github.com/opsybot/opsybot/internal/repository/workspace"
	"github.com/opsybot/opsybot/internal/service"
	"github.com/opsybot/opsybot/internal/service/auth"
	"github.com/opsybot/opsybot/internal/service/channels"
	"github.com/opsybot/opsybot/internal/service/members"
	"github.com/opsybot/opsybot/internal/service/references"
	"github.com/opsybot/opsybot/internal/service/teams"
	"github.com/opsybot/opsybot/internal/service/users"
	"github.com/opsybot/opsybot/internal/service/workspaces"
)

var repositoryProviders = wire.NewSet(
	transactor.New,
	lock.New,
	user.New,
	workspace.New,
	member.New,
	session.New,
	policy.New,
	invite.New,
	team.New,
	audit.New,
	mailer.New,
	password_reset.New,
	recovery_code.New,
	channel.New,
	pending.New,
)

var serviceProviders = wire.NewSet(
	wire.Value([]service.ReferenceSource(nil)),
	auth.New,
	workspaces.New,
	members.New,
	references.New,
	users.New,
	channels.New,
	teams.New,
)
