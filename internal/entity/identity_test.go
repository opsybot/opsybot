package entity

import "testing"

func TestIdentityScopePermits(t *testing.T) {
	session := Identity{Kind: IdentityKindSession, UserID: "u1"}
	auditKey := Identity{Kind: IdentityKindAPIKey, KeyKind: KeyKindWorkspace, Scopes: []Scope{ScopeAuditRead}}
	configKey := Identity{Kind: IdentityKindAPIKey, KeyKind: KeyKindPersonal, Scopes: []Scope{ScopeConfigWrite}}
	emptyKey := Identity{Kind: IdentityKindAPIKey}

	cases := []struct {
		name string
		id   Identity
		obj  PolicyObject
		act  PolicyAction
		want bool
	}{
		{"session bypasses scope on unmapped object", session, PolicyObjectMembers, PolicyActionWrite, true},
		{"session bypasses scope on mapped object", session, PolicyObjectAudit, PolicyActionRead, true},
		{"api key denied on unmapped object", auditKey, PolicyObjectMembers, PolicyActionRead, false},
		{"api key allowed on mapped object it holds", auditKey, PolicyObjectAudit, PolicyActionRead, true},
		{"api key denied on mapped object it lacks", auditKey, PolicyObjectSettings, PolicyActionWrite, false},
		{"config key allowed on settings write", configKey, PolicyObjectSettings, PolicyActionWrite, true},
		{"config key denied on audit read", configKey, PolicyObjectAudit, PolicyActionRead, false},
		{"scopeless key denied everywhere", emptyKey, PolicyObjectAudit, PolicyActionRead, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.id.ScopePermits(c.obj, c.act); got != c.want {
				t.Errorf("ScopePermits(%s, %s) = %v, want %v", c.obj, c.act, got, c.want)
			}
		})
	}
}

func TestIdentitySubject(t *testing.T) {
	workspaceKey := Identity{Kind: IdentityKindAPIKey, KeyKind: KeyKindWorkspace, WorkspaceID: "ws1"}
	if got := workspaceKey.Subject(); got != "wsagent:ws1" {
		t.Errorf("workspace key subject = %q, want wsagent:ws1", got)
	}
	personalKey := Identity{Kind: IdentityKindAPIKey, KeyKind: KeyKindPersonal, UserID: "u1", WorkspaceID: "ws1"}
	if got := personalKey.Subject(); got != "user:u1" {
		t.Errorf("personal key subject = %q, want user:u1", got)
	}
	session := Identity{Kind: IdentityKindSession, UserID: "u1"}
	if got := session.Subject(); got != "user:u1" {
		t.Errorf("session subject = %q, want user:u1", got)
	}
}
