package internal

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/repository/api_key"
	"github.com/opsybot/opsybot/internal/repository/audit"
	"github.com/opsybot/opsybot/internal/repository/channel"
	"github.com/opsybot/opsybot/internal/repository/invite"
	"github.com/opsybot/opsybot/internal/repository/lock"
	"github.com/opsybot/opsybot/internal/repository/mailer"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/password_reset"
	"github.com/opsybot/opsybot/internal/repository/pending"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/ratelimit"
	"github.com/opsybot/opsybot/internal/repository/recovery_code"
	"github.com/opsybot/opsybot/internal/repository/schedule"
	"github.com/opsybot/opsybot/internal/repository/session"
	"github.com/opsybot/opsybot/internal/repository/sso_connection"
	"github.com/opsybot/opsybot/internal/repository/sso_state"
	"github.com/opsybot/opsybot/internal/repository/team"
	"github.com/opsybot/opsybot/internal/repository/transactor"
	"github.com/opsybot/opsybot/internal/repository/user"
	"github.com/opsybot/opsybot/internal/repository/user_identity"
	"github.com/opsybot/opsybot/internal/repository/workspace"
	"github.com/opsybot/opsybot/internal/service"
	"github.com/opsybot/opsybot/internal/service/apikeys"
	"github.com/opsybot/opsybot/internal/service/audits"
	"github.com/opsybot/opsybot/internal/service/auth"
	"github.com/opsybot/opsybot/internal/service/channels"
	"github.com/opsybot/opsybot/internal/service/members"
	"github.com/opsybot/opsybot/internal/service/ratelimiter"
	"github.com/opsybot/opsybot/internal/service/references"
	"github.com/opsybot/opsybot/internal/service/schedules"
	"github.com/opsybot/opsybot/internal/service/sso"
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
	api_key.New,
	sso_connection.New,
	user_identity.New,
	sso_state.New,
	ratelimit.New,
	audit.New,
	mailer.New,
	password_reset.New,
	recovery_code.New,
	channel.New,
	pending.New,
	schedule.New,
)

var serviceProviders = wire.NewSet(
	scheduleReferenceSources,
	auth.New,
	workspaces.New,
	members.New,
	references.New,
	users.New,
	channels.New,
	teams.New,
	schedules.New,
	apikeys.New,
	audits.New,
	sso.New,
	ratelimiter.New,
)

func scheduleReferenceSources(schedules service.Schedules) []service.ReferenceSource {
	src, ok := schedules.(service.ReferenceSource)
	if !ok {
		return nil
	}
	return []service.ReferenceSource{src}
}
