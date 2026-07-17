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
	ActionTeamCreated       = "team.created"
	ActionTeamUpdated       = "team.updated"
	ActionTeamArchived      = "team.archived"
	ActionTeamUnarchived    = "team.unarchived"
	ActionKeyCreated        = "key.created"
	ActionKeyRevoked        = "key.revoked"
	ActionSettingsUpdated   = "settings.updated"
	ActionSSOUpdated        = "sso.updated"
	ActionChannelAdded      = "channel.added"
	ActionChannelRemoved    = "channel.removed"
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
