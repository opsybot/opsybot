import type {
	ApiKey,
	AuditEntry,
	ConfigDiff,
	CustomField,
	KeyKind,
	Member,
	Role,
	Team,
	WorkspaceSettings
} from '$lib/admin';
import { slugify, uid } from '$lib/admin';
import { scenario } from './fixtures';

const REFERENCES = [
	{ id: 'ref1', kind: 'schedule', icon: 'calendar-clock', label: 'payments-primary · layer 1 rotation', detail: 'Weekly rotation participant' },
	{ id: 'ref2', kind: 'policy', icon: 'arrow-up-right', label: 'frontend-daytime · escalation step 2', detail: 'Direct target, all-at-once' },
	{ id: 'ref3', kind: 'followup', icon: 'list-checks', label: 'INC-2481 follow-up owner', detail: '"Add canary stage to payments-api deploys"' }
];

const DEFAULT_SETTINGS = (): WorkspaceSettings => ({
	name: 'Acme Corp',
	timezone: 'UTC',
	severities: [
		{ id: 'SEV1', def: 'Customer-facing outage or data loss. All hands, page immediately.' },
		{ id: 'SEV2', def: 'Major degradation for many customers. Page the on-call now.' },
		{ id: 'SEV3', def: 'Partial or contained impact. Fix during working hours.' },
		{ id: 'SEV4', def: 'Minor issue, no customer impact yet. Track it.' }
	],
	postmortemThreshold: 'SEV2',
	fields: [
		{ id: 'f1', name: 'Customer impact', type: 'text' },
		{ id: 'f2', name: 'Detected by', type: 'select', options: 'synthetic monitoring, alert, customer report, internal' },
		{ id: 'f3', name: 'Regions affected', type: 'multi-select', options: 'eu-west-1, us-east-1, ap-southeast-2' }
	],
	cadence: { SEV1: '15 min', SEV2: '30 min', SEV3: '2 h', SEV4: 'none' },
	sso: { mode: 'oidc', issuer: 'https://sso.acme.dev/realms/acme', clientId: 'opsybot' },
	retention: { alerts: '1 year', incidents: 'forever', audit: '2 years' }
});

const CONFIG_DIFF: ConfigDiff = {
	created: [
		{ path: 'schedules/security-oncall', note: 'new schedule, 2 layers' },
		{ path: 'policies/security-primary', note: 'new escalation policy' }
	],
	changed: [
		{ path: 'schedules/payments-primary', note: 'handover 09:00 → 08:00 UTC' },
		{ path: 'routing/rule[2]', note: 'adds condition labels.env is prod' }
	],
	decision: [
		{ path: 'policies/frontend-daytime', note: 'references user tom@acme.dev — deactivated. Pick a replacement or skip.' }
	],
	skipped: [{ path: 'statuspages/status.acme.dev', note: 'identical to current config' }]
};

function seed() {
	const members: Member[] = [
		{ id: 'u1', name: 'Maya Chen', email: 'maya@acme.dev', role: 'admin', twoFactor: true, auth: 'SSO (OIDC)', lastActive: '2 m ago', references: [], deactivated: false },
		{ id: 'u2', name: 'Priya Nair', email: 'priya@acme.dev', role: 'admin', twoFactor: true, auth: 'SSO (OIDC)', lastActive: '18 m ago', references: [], deactivated: false },
		{ id: 'u3', name: 'Marcus Lee', email: 'marcus@acme.dev', role: 'responder', twoFactor: true, auth: 'password', lastActive: '1 h ago', references: REFERENCES, deactivated: false },
		{ id: 'u4', name: 'Dev Patel', email: 'dev@acme.dev', role: 'responder', twoFactor: false, auth: 'password', lastActive: '3 h ago', references: [], deactivated: false },
		{ id: 'u5', name: 'Sana Ito', email: 'sana@acme.dev', role: 'responder', twoFactor: true, auth: 'SSO (OIDC)', lastActive: 'yesterday', references: [], deactivated: false },
		{ id: 'u6', name: 'Jordan Okafor', email: 'jordan@acme.dev', role: 'viewer', twoFactor: false, auth: 'invited', lastActive: 'never', references: [], deactivated: false }
	];
	const teams: Team[] = [
		{ id: 'payments', name: 'payments', members: ['Maya Chen', 'Marcus Lee'], schedules: ['payments-primary'], policies: ['payments-primary'], services: ['payments-api', 'events-worker'] },
		{ id: 'platform', name: 'platform', members: ['Priya Nair', 'Dev Patel'], schedules: ['platform-default'], policies: ['platform-default'], services: ['gateway', 'database'] },
		{ id: 'frontend', name: 'frontend', members: ['Sana Ito', 'Dev Patel'], schedules: ['frontend-daytime'], policies: ['frontend-daytime'], services: ['checkout-web', 'edge'] }
	];
	const personalKeys: ApiKey[] = [
		{ id: 'k1', name: 'laptop-cli', scopes: ['incidents:read', 'alerts:write'], last: '2026-07-11 08:12 UTC', created: '2026-05-02' },
		{ id: 'k2', name: 'homelab-scripts', scopes: ['alerts:write'], last: 'never', created: '2026-06-20' }
	];
	const workspaceKeys: ApiKey[] = [
		{ id: 'k3', name: 'terraform-provider', scopes: ['config:write', 'schedules:write', 'policies:write'], last: '2026-07-10 22:00 UTC', created: '2026-03-14' },
		{ id: 'k4', name: 'grafana-annotations', scopes: ['incidents:read'], last: '2026-07-11 09:41 UTC', created: '2026-04-08' }
	];
	const audit: AuditEntry[] = [
		{ id: 'a1', at: '2026-07-11 09:52:18 UTC', actor: 'Maya Chen', action: 'alert.escalate', target: 'al-6 → step 3', ip: '10.0.2.14' },
		{ id: 'a2', at: '2026-07-11 09:38:02 UTC', actor: 'Priya Nair', action: 'statuspage.update.publish', target: 'status.acme.dev · INC-2481', ip: '10.0.2.31' },
		{ id: 'a3', at: '2026-07-11 09:14:02 UTC', actor: 'Maya Chen', action: 'incident.declare', target: 'INC-2481 (SEV1)', ip: '10.0.2.14' },
		{ id: 'a4', at: '2026-07-11 08:55:40 UTC', actor: 'Dev Patel', action: 'auth.login', target: 'password + 2FA', ip: '84.121.9.77' },
		{ id: 'a5', at: '2026-07-11 08:02:19 UTC', actor: 'system', action: 'integration.parse_failure', target: 'legacy-nagios', ip: '—' },
		{ id: 'a6', at: '2026-07-10 22:00:11 UTC', actor: 'terraform-provider', action: 'config.apply', target: '3 schedules, 2 policies', ip: '10.0.8.2' },
		{ id: 'a7', at: '2026-07-10 18:40:56 UTC', actor: 'Sana Ito', action: 'auth.login.failed', target: 'wrong password ×2', ip: '203.0.113.9' },
		{ id: 'a8', at: '2026-07-10 15:22:03 UTC', actor: 'Priya Nair', action: 'member.role.change', target: 'Jordan Okafor → viewer', ip: '10.0.2.31' }
	];
	return { members, teams, personalKeys, workspaceKeys, audit, settings: DEFAULT_SETTINGS() };
}

const store = seed();

const state = scenario();
if (state === 'empty') {
	store.members = store.members.filter((member) => member.email === 'maya@acme.dev').map((member) => ({ ...member, references: [] }));
	store.teams = [];
	store.personalKeys = [];
	store.workspaceKeys = [];
	store.audit = [];
	store.settings = DEFAULT_SETTINGS();
}
if (state === 'degraded') {
	const sana = store.members.find((member) => member.email === 'sana@acme.dev');
	if (sana) sana.twoFactor = false;
	store.audit.unshift({ id: uid('a'), at: '2026-07-15 03:11:02 UTC', actor: 'unknown', action: 'auth.login.failed', target: 'sam@sso.acme.dev · IP not recognised', ip: '198.51.100.9' });
}

function nowStamp(): string {
	return `${new Date().toISOString().replace('T', ' ').slice(0, 19)} UTC`;
}

function append(action: string, target: string, actor = 'Maya Chen') {
	store.audit.unshift({ id: uid('a'), at: nowStamp(), actor, action, target, ip: '10.0.2.14' });
}

export function listMembers(): Member[] {
	return store.members;
}

export function inviteMember(email: string, role: Role): Member | { error: string } {
	if (store.members.some((member) => member.email.toLowerCase() === email.toLowerCase()))
		return { error: 'That email is already a member.' };
	const member: Member = {
		id: uid('u'),
		name: email.split('@')[0],
		email,
		role,
		twoFactor: false,
		auth: 'invited',
		lastActive: 'never',
		references: [],
		deactivated: false
	};
	store.members.push(member);
	append('member.invited', `${email} (${role})`);
	return member;
}

export function changeRole(id: string, role: Role): boolean {
	const member = store.members.find((entry) => entry.id === id);
	if (!member || member.deactivated) return false;
	// Keep at least one active admin
	const admins = store.members.filter((entry) => entry.role === 'admin' && !entry.deactivated).length;
	if (member.role === 'admin' && role !== 'admin' && admins <= 1) return false;
	if (member.role === role) return true;
	member.role = role;
	append('member.role.change', `${member.name} → ${role}`);
	return true;
}

export function deactivateMember(id: string, replacements: Record<string, string>): boolean {
	const member = store.members.find((entry) => entry.id === id);
	if (!member || member.deactivated) return false;
	// Last-admin guard, same as changeRole
	if (member.role === 'admin' && store.members.filter((entry) => entry.role === 'admin' && !entry.deactivated).length <= 1)
		return false;
	const valid = new Set(store.members.filter((entry) => !entry.deactivated && entry.id !== id).map((entry) => entry.name));
	if (member.references.length && !member.references.every((ref) => valid.has(replacements[ref.id]))) return false;
	for (const ref of member.references) append('reference.reassign', `${ref.label} → ${replacements[ref.id]}`);
	member.references = [];
	member.deactivated = true;
	append('member.deactivated', member.name);
	return true;
}

export function reactivateMember(id: string): boolean {
	const member = store.members.find((entry) => entry.id === id);
	if (!member?.deactivated) return false;
	member.deactivated = false;
	append('member.reactivated', member.name);
	return true;
}

const RESERVED = new Set(['teams', 'new', 'keys', 'audit', 'settings', 'config']);

export function listTeams(): Team[] {
	return store.teams;
}

export function getTeam(id: string): Team | undefined {
	return store.teams.find((team) => team.id === id);
}

function uniqueSlug(name: string): string {
	const base = slugify(name) || 'team';
	let slug = base;
	let n = 2;
	while (RESERVED.has(slug) || store.teams.some((team) => team.id === slug)) slug = `${base}-${n++}`;
	return slug;
}

export function createTeam(name: string, members: string[]): Team {
	const team: Team = { id: uniqueSlug(name), name, members, schedules: [], policies: [], services: [] };
	store.teams.push(team);
	append('team.created', name);
	return team;
}

export function updateTeam(id: string, name: string, members: string[]): boolean {
	const team = getTeam(id);
	if (!team) return false;
	team.name = name;
	team.members = members;
	append('team.updated', name);
	return true;
}

export function listKeys(): { personal: ApiKey[]; workspace: ApiKey[] } {
	return { personal: store.personalKeys, workspace: store.workspaceKeys };
}

function hex(length: number): string {
	let out = '';
	while (out.length < length) out += Math.floor(Math.random() * 16).toString(16);
	return out.slice(0, length);
}

// Secret is returned once and never stored
export function createKey(name: string, scopes: string[], kind: KeyKind): { secret: string } {
	const secret = `osk_${kind.slice(0, 2)}_${hex(24)}`;
	const key: ApiKey = { id: uid('k'), name, scopes, last: 'never', created: nowStamp().slice(0, 10) };
	store[kind === 'personal' ? 'personalKeys' : 'workspaceKeys'].push(key);
	append('key.created', `${name} (${kind})`);
	return { secret };
}

export function revokeKey(id: string): boolean {
	for (const kind of ['personalKeys', 'workspaceKeys'] as const) {
		const index = store[kind].findIndex((key) => key.id === id);
		if (index >= 0) {
			const [removed] = store[kind].splice(index, 1);
			append('key.revoked', removed.name);
			return true;
		}
	}
	return false;
}

export function listAudit(): AuditEntry[] {
	return store.audit;
}

export function getSettings(): WorkspaceSettings {
	return store.settings;
}

export function saveSettings(next: WorkspaceSettings): void {
	store.settings = { ...next, fields: store.settings.fields };
	append('settings.updated', 'workspace settings');
}

export function setThreshold(threshold: string): boolean {
	if (!['SEV1', 'SEV2', 'SEV3'].includes(threshold)) return false;
	store.settings.postmortemThreshold = threshold;
	append('settings.postmortem_threshold', `${threshold} and above`);
	return true;
}

export function addField(name: string, type: string): CustomField {
	const field: CustomField = { id: uid('f'), name, type };
	store.settings.fields.push(field);
	append('settings.field_added', name);
	return field;
}

export function removeField(id: string): boolean {
	const index = store.settings.fields.findIndex((field) => field.id === id);
	if (index < 0) return false;
	const [removed] = store.settings.fields.splice(index, 1);
	append('settings.field_removed', removed.name);
	return true;
}

export function getDiff(): ConfigDiff {
	return CONFIG_DIFF;
}

export function exportYaml(): string {
	const settings = store.settings;
	const lines = [
		'# Opsybot workspace configuration',
		`workspace: ${settings.name}`,
		`timezone: ${settings.timezone}`,
		'severities:',
		...settings.severities.map((sev) => `  - id: ${sev.id}\n    definition: "${sev.def}"`),
		`postmortem_required_for: ${settings.postmortemThreshold}`,
		'teams:',
		...store.teams.map((team) => `  - name: ${team.name}\n    members: [${team.members.join(', ')}]`),
		'custom_fields:',
		...settings.fields.map((field) => `  - name: "${field.name}"\n    type: ${field.type}`)
	];
	return lines.join('\n') + '\n';
}

export function applyImport(decision: string): boolean {
	if (decision !== 'replace' && decision !== 'skip') return false;
	append('config.apply', `2 created, 2 changed, 1 ${decision === 'skip' ? 'skipped' : 'reassigned'}, 1 unchanged`);
	return true;
}
