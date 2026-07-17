package internal

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/repository/lock"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/session"
	"github.com/opsybot/opsybot/internal/repository/user"
	"github.com/opsybot/opsybot/internal/repository/workspace"
	"github.com/opsybot/opsybot/internal/service/auth"
	"github.com/opsybot/opsybot/internal/service/members"
	"github.com/opsybot/opsybot/internal/service/workspaces"
)

var repositoryProviders = wire.NewSet(
	lock.New,
	user.New,
	workspace.New,
	member.New,
	session.New,
	policy.New,
)

var serviceProviders = wire.NewSet(
	auth.New,
	workspaces.New,
	members.New,
)
