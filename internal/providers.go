package internal

import (
	"github.com/goforj/wire"

	"github.com/opsybot/opsybot/internal/cron"
	"github.com/opsybot/opsybot/internal/repository/action_token"
	"github.com/opsybot/opsybot/internal/repository/alert"
	"github.com/opsybot/opsybot/internal/repository/alert_monitor"
	"github.com/opsybot/opsybot/internal/repository/alert_route"
	"github.com/opsybot/opsybot/internal/repository/alert_source"
	"github.com/opsybot/opsybot/internal/repository/api_key"
	"github.com/opsybot/opsybot/internal/repository/audit"
	"github.com/opsybot/opsybot/internal/repository/channel"
	"github.com/opsybot/opsybot/internal/repository/channel_verification"
	"github.com/opsybot/opsybot/internal/repository/chat_connection"
	"github.com/opsybot/opsybot/internal/repository/chat_courier"
	"github.com/opsybot/opsybot/internal/repository/chat_identity"
	"github.com/opsybot/opsybot/internal/repository/chat_oauth_state"
	"github.com/opsybot/opsybot/internal/repository/escalation_policy"
	"github.com/opsybot/opsybot/internal/repository/escalation_run"
	"github.com/opsybot/opsybot/internal/repository/incident"
	"github.com/opsybot/opsybot/internal/repository/incident_field_def"
	"github.com/opsybot/opsybot/internal/repository/incident_severity"
	"github.com/opsybot/opsybot/internal/repository/ingest_event"
	"github.com/opsybot/opsybot/internal/repository/invite"
	"github.com/opsybot/opsybot/internal/repository/lock"
	"github.com/opsybot/opsybot/internal/repository/mailer"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/notification_rule"
	"github.com/opsybot/opsybot/internal/repository/notification_run"
	"github.com/opsybot/opsybot/internal/repository/ntfy"
	"github.com/opsybot/opsybot/internal/repository/pager"
	"github.com/opsybot/opsybot/internal/repository/password_reset"
	"github.com/opsybot/opsybot/internal/repository/pending"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/ratelimit"
	"github.com/opsybot/opsybot/internal/repository/recovery_code"
	"github.com/opsybot/opsybot/internal/repository/schedule"
	servicerepo "github.com/opsybot/opsybot/internal/repository/service"
	"github.com/opsybot/opsybot/internal/repository/session"
	"github.com/opsybot/opsybot/internal/repository/silence"
	"github.com/opsybot/opsybot/internal/repository/sso_connection"
	"github.com/opsybot/opsybot/internal/repository/sso_state"
	"github.com/opsybot/opsybot/internal/repository/team"
	"github.com/opsybot/opsybot/internal/repository/transactor"
	"github.com/opsybot/opsybot/internal/repository/user"
	"github.com/opsybot/opsybot/internal/repository/user_identity"
	"github.com/opsybot/opsybot/internal/repository/workspace"
	"github.com/opsybot/opsybot/internal/service"
	"github.com/opsybot/opsybot/internal/service/actions"
	"github.com/opsybot/opsybot/internal/service/alert_monitors"
	"github.com/opsybot/opsybot/internal/service/alert_routes"
	"github.com/opsybot/opsybot/internal/service/alert_sources"
	"github.com/opsybot/opsybot/internal/service/alerts"
	"github.com/opsybot/opsybot/internal/service/apikeys"
	"github.com/opsybot/opsybot/internal/service/audits"
	"github.com/opsybot/opsybot/internal/service/auth"
	"github.com/opsybot/opsybot/internal/service/channels"
	"github.com/opsybot/opsybot/internal/service/chats"
	"github.com/opsybot/opsybot/internal/service/escalations"
	"github.com/opsybot/opsybot/internal/service/incidents"
	"github.com/opsybot/opsybot/internal/service/ingest"
	"github.com/opsybot/opsybot/internal/service/interactions"
	"github.com/opsybot/opsybot/internal/service/members"
	"github.com/opsybot/opsybot/internal/service/notification_rules"
	"github.com/opsybot/opsybot/internal/service/notifications"
	"github.com/opsybot/opsybot/internal/service/notifier"
	"github.com/opsybot/opsybot/internal/service/ratelimiter"
	"github.com/opsybot/opsybot/internal/service/references"
	"github.com/opsybot/opsybot/internal/service/schedules"
	serviceservices "github.com/opsybot/opsybot/internal/service/services"
	"github.com/opsybot/opsybot/internal/service/silences"
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
	alert_source.New,
	alert.New,
	ingest_event.New,
	alert_route.New,
	alert_monitor.New,
	silence.New,
	escalation_policy.New,
	escalation_run.New,
	notification_rule.New,
	notification_run.New,
	channel_verification.New,
	chat_connection.New,
	chat_identity.New,
	chat_courier.New,
	chat_oauth_state.New,
	action_token.New,
	pager.New,
	ntfy.New,
	servicerepo.New,
	incident.New,
	incident_severity.New,
	incident_field_def.New,
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
	alert_sources.New,
	alerts.New,
	ingest.New,
	alert_routes.New,
	alert_monitors.New,
	silences.New,
	notifier.New,
	notification_rules.New,
	notifications.New,
	actions.New,
	interactions.New,
	chats.New,
	escalations.New,
	serviceservices.New,
	incidents.New,
)

var cronProviders = wire.NewSet(
	cron.NewHeartbeatSweep,
	cron.NewAlertAutoResolve,
	cron.NewIngestRetention,
	cron.NewEscalationSweep,
	cron.NewNotificationSweep,
)

func scheduleReferenceSources(schedules service.Schedules, escalations service.Escalations) []service.ReferenceSource {
	out := make([]service.ReferenceSource, 0, 2)
	if src, ok := schedules.(service.ReferenceSource); ok {
		out = append(out, src)
	}
	if src, ok := escalations.(service.ReferenceSource); ok {
		out = append(out, src)
	}
	return out
}
