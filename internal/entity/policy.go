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
)

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
		}
	case RoleMember:
		return []PolicyRule{
			{PolicyObjectMembers, PolicyActionRead},
			{PolicyObjectTeams, PolicyActionRead},
			{PolicyObjectSettings, PolicyActionRead},
			{PolicyObjectPersonalKeys, PolicyActionRead}, {PolicyObjectPersonalKeys, PolicyActionWrite},
			{PolicyObjectChannels, PolicyActionRead}, {PolicyObjectChannels, PolicyActionWrite},
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
	default:
		return "", false
	}
}
