package entity

type PolicyObject string

const (
	PolicyObjectMembers      PolicyObject = "members"
	PolicyObjectTeams        PolicyObject = "teams"
	PolicyObjectSettings     PolicyObject = "settings"
	PolicyObjectSSO          PolicyObject = "sso"
	PolicyObjectKeys         PolicyObject = "keys"
	PolicyObjectPersonalKeys PolicyObject = "personal_keys"
	PolicyObjectChannels     PolicyObject = "channels"
	PolicyObjectAudit        PolicyObject = "audit"
	PolicyObjectSchedules    PolicyObject = "schedules"
	PolicyObjectAlerts       PolicyObject = "alerts"
	PolicyObjectAlertSources PolicyObject = "alert_sources"
	PolicyObjectPolicies     PolicyObject = "policies"
	PolicyObjectChat         PolicyObject = "chat"
)

const PolicyRefMaxLength = 40

type PolicyAction string

const (
	PolicyActionRead  PolicyAction = "read"
	PolicyActionWrite PolicyAction = "write"
)

type PolicyRule struct {
	Object PolicyObject
	Action PolicyAction
}

func RolePolicies(role Role) []PolicyRule {
	switch role {
	case RoleAdmin:
		return []PolicyRule{
			{PolicyObjectMembers, PolicyActionRead}, {PolicyObjectMembers, PolicyActionWrite},
			{PolicyObjectTeams, PolicyActionRead}, {PolicyObjectTeams, PolicyActionWrite},
			{PolicyObjectSettings, PolicyActionRead}, {PolicyObjectSettings, PolicyActionWrite},
			{PolicyObjectSSO, PolicyActionRead}, {PolicyObjectSSO, PolicyActionWrite},
			{PolicyObjectKeys, PolicyActionRead}, {PolicyObjectKeys, PolicyActionWrite},
			{PolicyObjectPersonalKeys, PolicyActionRead}, {PolicyObjectPersonalKeys, PolicyActionWrite},
			{PolicyObjectChannels, PolicyActionRead}, {PolicyObjectChannels, PolicyActionWrite},
			{PolicyObjectAudit, PolicyActionRead},
			{PolicyObjectSchedules, PolicyActionRead}, {PolicyObjectSchedules, PolicyActionWrite},
			{PolicyObjectAlerts, PolicyActionRead}, {PolicyObjectAlerts, PolicyActionWrite},
			{PolicyObjectAlertSources, PolicyActionRead},
			{PolicyObjectPolicies, PolicyActionRead}, {PolicyObjectAlertSources, PolicyActionWrite},
			{PolicyObjectPolicies, PolicyActionRead}, {PolicyObjectPolicies, PolicyActionWrite},
			{PolicyObjectChat, PolicyActionRead}, {PolicyObjectChat, PolicyActionWrite},
		}
	case RoleMember:
		return []PolicyRule{
			{PolicyObjectMembers, PolicyActionRead},
			{PolicyObjectTeams, PolicyActionRead},
			{PolicyObjectSettings, PolicyActionRead},
			{PolicyObjectPersonalKeys, PolicyActionRead}, {PolicyObjectPersonalKeys, PolicyActionWrite},
			{PolicyObjectChannels, PolicyActionRead}, {PolicyObjectChannels, PolicyActionWrite},
			{PolicyObjectSchedules, PolicyActionRead},
			{PolicyObjectAlerts, PolicyActionRead}, {PolicyObjectAlerts, PolicyActionWrite},
			{PolicyObjectAlertSources, PolicyActionRead},
			{PolicyObjectPolicies, PolicyActionRead},
			{PolicyObjectChat, PolicyActionRead},
		}
	default:
		return nil
	}
}

func ScopeFor(obj PolicyObject, act PolicyAction) (Scope, bool) {
	switch {
	case obj == PolicyObjectAudit && act == PolicyActionRead:
		return ScopeAuditRead, true
	case obj == PolicyObjectSettings && act == PolicyActionWrite:
		return ScopeConfigWrite, true
	case obj == PolicyObjectSchedules && act == PolicyActionWrite:
		return ScopeSchedulesWrite, true
	case obj == PolicyObjectAlerts && act == PolicyActionRead:
		return ScopeAlertsRead, true
	case obj == PolicyObjectAlerts && act == PolicyActionWrite:
		return ScopeAlertsWrite, true
	case obj == PolicyObjectAlertSources && act == PolicyActionWrite:
		return ScopeAlertsWrite, true
	case obj == PolicyObjectPolicies && act == PolicyActionWrite:
		return ScopePoliciesWrite, true
	default:
		return "", false
	}
}
