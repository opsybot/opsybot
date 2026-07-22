package entity

import (
	"errors"
	"time"
)

var ErrAuditInvalidCursor = errors.New("audit invalid cursor")

type AuditActorType string

const (
	AuditActorUser    AuditActorType = "user"
	AuditActorAPIKey  AuditActorType = "api_key"
	AuditActorSystem  AuditActorType = "system"
	AuditActorUnknown AuditActorType = "unknown"
)

const (
	ActionInstanceSetup     = "instance.setup"
	ActionWorkspaceCreated  = "workspace.created"
	ActionMemberInvited     = "member.invited"
	ActionMemberJoined      = "member.joined"
	ActionMemberRoleChange  = "member.role.change"
	ActionMemberDeactivated = "member.deactivated"
	ActionMemberReactivated = "member.reactivated"
	ActionInviteRevoked     = "member.invite.revoked"
	ActionReferenceReassign = "reference.reassign"
	ActionAuthLogin         = "auth.login"
	ActionAuthLoginFailed   = "auth.login.failed"
	ActionAuthLogout        = "auth.logout"

	ActionPasswordResetRequested = "auth.password.reset.requested"
	ActionPasswordResetCompleted = "auth.password.reset.completed"
	ActionKeyUsed                = "key.used"

	ActionTeamCreated     = "team.created"
	ActionTeamUpdated     = "team.updated"
	ActionTeamArchived    = "team.archived"
	ActionTeamUnarchived  = "team.unarchived"
	ActionKeyCreated      = "key.created"
	ActionKeyRevoked      = "key.revoked"
	ActionSettingsUpdated = "settings.updated"
	ActionSSOUpdated      = "sso.updated"
	ActionChannelAdded    = "channel.added"
	ActionChannelRemoved  = "channel.removed"
	ActionChannelVerified = "channel.verified"
	ActionChannelTested   = "channel.tested"

	ActionNotificationRulesSaved = "notification_rules.saved"

	ActionScheduleCreated     = "schedule.created"
	ActionScheduleUpdated     = "schedule.updated"
	ActionScheduleDuplicated  = "schedule.duplicated"
	ActionScheduleArchived    = "schedule.archived"
	ActionScheduleUnarchived  = "schedule.unarchived"
	ActionSchedulePaused      = "schedule.paused"
	ActionScheduleResumed     = "schedule.resumed"
	ActionScheduleDeleted     = "schedule.deleted"
	ActionScheduleOverrideAdd = "schedule.override.added"
	ActionScheduleReassigned  = "schedule.reassigned"

	ActionAlertSourceCreated       = "alert_source.created"
	ActionAlertSourceUpdated       = "alert_source.updated"
	ActionAlertSourceDeleted       = "alert_source.deleted"
	ActionAlertSourcePaused        = "alert_source.paused"
	ActionAlertSourceResumed       = "alert_source.resumed"
	ActionAlertSourceSecretRotated = "alert_source.secret_rotated"
	ActionAlertSourceMappingSaved  = "alert_source.mapping_saved"
	ActionAlertAcknowledged        = "alert.acknowledged"
	ActionAlertResolved            = "alert.resolved"
	ActionAlertMonitorCreated      = "alert_monitor.created"
	ActionAlertMonitorUpdated      = "alert_monitor.updated"
	ActionAlertMonitorDeleted      = "alert_monitor.deleted"
	ActionAlertGroupRulesSaved     = "alert_group_rules.saved"
	ActionPolicyCreated            = "policy.created"
	ActionPolicyUpdated            = "policy.updated"
	ActionPolicyDeleted            = "policy.deleted"
	ActionAlertEscalated           = "alert.escalated"
	ActionEscalationWebhookCreated = "escalation_webhook.created"
	ActionEscalationWebhookUpdated = "escalation_webhook.updated"
	ActionEscalationWebhookDeleted = "escalation_webhook.deleted"
)

const (
	AuditDefaultLimit = 50
	AuditMaxLimit     = 200
	ActorLabelSystem  = "system"
	ActorLabelUnknown = "unknown"
)

type AuditEvent struct {
	ID          string
	WorkspaceID string
	At          time.Time
	ActorType   AuditActorType
	ActorUserID string
	ActorLabel  string
	Action      string
	Target      string
	IP          string
	Meta        map[string]string
}

type AuditFilter struct {
	Query        string
	ActorUserID  string
	ActionPrefix string
	Cursor       string
	Limit        int
}

type AuditPage struct {
	Events     []AuditEvent
	NextCursor string
}
