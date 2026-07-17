import type { Cookies } from '@sveltejs/kit';
import type {
	ApiKey,
	AuditEntry,
	ConfigDiff,
	CustomField,
	KeyKind,
	Member,
	MemberReference,
	Role,
	Team,
	WorkspaceSettings
} from '$lib/admin';
import { uid } from '$lib/admin';
import { api } from './api';

function ago(iso?: string | null): string {
	if (!iso) return 'never';
	const at = Date.parse(iso);
	if (Number.isNaN(at)) return 'never';
	const seconds = Math.max(0, Math.floor((Date.now() - at) / 1000));
	if (seconds < 60) return `${seconds}s ago`;
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes} m ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours} h ago`;
	const days = Math.floor(hours / 24);
	return days === 1 ? 'yesterday' : `${days} d ago`;
}

// ---- Members -------------------------------------------------------------

type MemberDTO = {
	userId: string;
	name: string;
	email: string;
	role: Role;
	status: 'invited' | 'active' | 'deactivated';
	twoFactor: boolean;
	authMethod: 'password' | 'sso' | 'invited';
	deactivated: boolean;
	lastActiveAt?: string;
	references?: { id: string; kind: string; icon?: string; label: string; detail: string }[];
};

function mapMember(dto: MemberDTO): Member {
	const auth = dto.authMethod === 'sso' ? 'SSO' : dto.authMethod === 'invited' ? 'invited' : 'password';
	const references: MemberReference[] = (dto.references ?? []).map((ref) => ({
		id: ref.id,
		kind: ref.kind,
		icon: ref.icon ?? '',
		label: ref.label,
		detail: ref.detail
	}));
	return {
		id: dto.userId,
		name: dto.name,
		email: dto.email,
		role: dto.role,
		twoFactor: dto.twoFactor,
		auth,
		lastActive: dto.status === 'invited' ? 'never' : ago(dto.lastActiveAt),
		references,
		deactivated: dto.deactivated
	};
}

export async function listMembers(cookies: Cookies, workspace: string): Promise<Member[]> {
	const res = await api.get<{ items: MemberDTO[] }>(`/workspaces/${workspace}/members`, cookies);
	return (res.data?.items ?? []).map(mapMember);
}

export async function inviteMember(
	cookies: Cookies,
	workspace: string,
	email: string,
	role: Role
): Promise<{ error?: string }> {
	const res = await api.post(`/workspaces/${workspace}/members/invites`, cookies, {
		body: { email, role }
	});
	return res.ok ? {} : { error: res.problem?.detail ?? 'Could not invite that person.' };
}

export async function changeRole(
	cookies: Cookies,
	workspace: string,
	userId: string,
	role: Role
): Promise<boolean> {
	const res = await api.patch(`/workspaces/${workspace}/members/${userId}/role`, cookies, {
		body: { role }
	});
	return res.ok;
}

export async function deactivateMember(
	cookies: Cookies,
	workspace: string,
	userId: string,
	replacements: Record<string, string>
): Promise<boolean> {
	const res = await api.post(`/workspaces/${workspace}/members/${userId}/deactivate`, cookies, {
		body: { replacements }
	});
	return res.ok;
}

export async function reactivateMember(
	cookies: Cookies,
	workspace: string,
	userId: string
): Promise<boolean> {
	const res = await api.post(`/workspaces/${workspace}/members/${userId}/reactivate`, cookies, {});
	return res.ok;
}

// ---- Teams ---------------------------------------------------------------

type TeamDTO = { id: string; slug: string; name: string; memberIds: string[]; archived: boolean };

async function memberNameIndex(cookies: Cookies, workspace: string) {
	const members = await listMembers(cookies, workspace);
	const byId = new Map(members.map((member) => [member.id, member.name]));
	const byName = new Map(members.map((member) => [member.name, member.id]));
	return { byId, byName };
}

function mapTeam(dto: TeamDTO, byId: Map<string, string>): Team {
	return {
		id: dto.slug,
		name: dto.name,
		members: dto.memberIds.map((id) => byId.get(id) ?? id),
		schedules: [],
		policies: [],
		services: []
	};
}

export async function listTeams(cookies: Cookies, workspace: string): Promise<Team[]> {
	const [res, { byId }] = await Promise.all([
		api.get<{ items: TeamDTO[] }>(`/workspaces/${workspace}/teams`, cookies),
		memberNameIndex(cookies, workspace)
	]);
	return (res.data?.items ?? []).map((team) => mapTeam(team, byId));
}

export async function getTeam(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<Team | undefined> {
	const [res, { byId }] = await Promise.all([
		api.get<TeamDTO>(`/workspaces/${workspace}/teams/${slug}`, cookies),
		memberNameIndex(cookies, workspace)
	]);
	return res.ok && res.data ? mapTeam(res.data, byId) : undefined;
}

export async function createTeam(
	cookies: Cookies,
	workspace: string,
	name: string,
	memberNames: string[]
): Promise<{ id?: string; error?: string }> {
	const { byName } = await memberNameIndex(cookies, workspace);
	const memberIds = memberNames.map((n) => byName.get(n)).filter((id): id is string => Boolean(id));
	const res = await api.post<TeamDTO>(`/workspaces/${workspace}/teams`, cookies, {
		body: { name, memberIds }
	});
	if (!res.ok || !res.data) return { error: res.problem?.detail ?? 'Could not create the team.' };
	return { id: res.data.slug };
}

export async function updateTeam(
	cookies: Cookies,
	workspace: string,
	slug: string,
	name: string,
	memberNames: string[]
): Promise<boolean> {
	const { byName } = await memberNameIndex(cookies, workspace);
	const memberIds = memberNames.map((n) => byName.get(n)).filter((id): id is string => Boolean(id));
	const res = await api.patch(`/workspaces/${workspace}/teams/${slug}`, cookies, {
		body: { name, memberIds }
	});
	return res.ok;
}

// ---- API keys ------------------------------------------------------------

type ApiKeyDTO = {
	id: string;
	name: string;
	kind: KeyKind;
	scopes: string[];
	hint: string;
	createdAt: string;
	lastUsedAt?: string;
};

function mapKey(dto: ApiKeyDTO): ApiKey {
	return {
		id: dto.id,
		name: dto.name,
		scopes: dto.scopes,
		last: dto.lastUsedAt ? ago(dto.lastUsedAt) : 'never',
		created: dto.createdAt.slice(0, 10)
	};
}

export async function listKeys(
	cookies: Cookies,
	workspace: string
): Promise<{ personal: ApiKey[]; workspace: ApiKey[] }> {
	const res = await api.get<{ personal: ApiKeyDTO[]; workspace: ApiKeyDTO[] }>(
		`/workspaces/${workspace}/keys`,
		cookies
	);
	return {
		personal: (res.data?.personal ?? []).map(mapKey),
		workspace: (res.data?.workspace ?? []).map(mapKey)
	};
}

export async function createKey(
	cookies: Cookies,
	workspace: string,
	name: string,
	scopes: string[],
	kind: KeyKind
): Promise<{ secret?: string; error?: string }> {
	const res = await api.post<{ secret: string }>(`/workspaces/${workspace}/keys`, cookies, {
		body: { name, kind, scopes }
	});
	if (!res.ok || !res.data) return { error: res.problem?.detail ?? 'Could not create the key.' };
	return { secret: res.data.secret };
}

export async function revokeKey(cookies: Cookies, workspace: string, id: string): Promise<boolean> {
	const res = await api.del(`/workspaces/${workspace}/keys/${id}`, cookies);
	return res.ok;
}

// ---- Audit ---------------------------------------------------------------

type AuditDTO = { id: string; at: string; actor: string; action: string; target: string; ip: string };

export async function listAudit(cookies: Cookies, workspace: string): Promise<AuditEntry[]> {
	const res = await api.get<{ items: AuditDTO[] }>(`/workspaces/${workspace}/audit?limit=100`, cookies);
	return (res.data?.items ?? []).map((entry) => ({
		id: entry.id,
		at: entry.at.replace('T', ' ').slice(0, 19) + ' UTC',
		actor: entry.actor || 'system',
		action: entry.action,
		target: entry.target,
		ip: entry.ip || '—'
	}));
}

// ---- Workspace settings (incident config is a deferred domain: fixture) ----

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
	sso: { mode: 'oidc', issuer: '', clientId: '' },
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

let settings = DEFAULT_SETTINGS();

type SsoConfigDTO = { mode: 'oidc' | 'saml'; issuer: string; clientId: string };

export async function getSettings(cookies: Cookies, workspace: string): Promise<WorkspaceSettings> {
	const res = await api.get<SsoConfigDTO>(`/workspaces/${workspace}/sso`, cookies);
	if (res.ok && res.data) {
		settings.sso = { mode: res.data.mode === 'saml' ? 'saml' : 'oidc', issuer: res.data.issuer, clientId: res.data.clientId };
	}
	return settings;
}

export async function saveSettings(
	cookies: Cookies,
	workspace: string,
	next: WorkspaceSettings
): Promise<void> {
	settings = { ...next, fields: settings.fields };
	await api.put(`/workspaces/${workspace}/sso`, cookies, {
		body: {
			mode: next.sso.mode,
			issuer: next.sso.issuer,
			clientId: next.sso.clientId,
			enabled: true,
			required: false,
			jitProvisioning: false
		}
	});
}

export function setThreshold(threshold: string): boolean {
	if (!['SEV1', 'SEV2', 'SEV3'].includes(threshold)) return false;
	settings.postmortemThreshold = threshold;
	return true;
}

export function addField(name: string, type: string): CustomField {
	const field: CustomField = { id: uid('f'), name, type };
	settings.fields.push(field);
	return field;
}

export function removeField(id: string): boolean {
	const index = settings.fields.findIndex((field) => field.id === id);
	if (index < 0) return false;
	settings.fields.splice(index, 1);
	return true;
}

export function getDiff(): ConfigDiff {
	return CONFIG_DIFF;
}

export function exportYaml(): string {
	const lines = [
		'# Opsybot workspace configuration',
		`workspace: ${settings.name}`,
		`timezone: ${settings.timezone}`,
		'severities:',
		...settings.severities.map((sev) => `  - id: ${sev.id}\n    definition: "${sev.def}"`),
		`postmortem_required_for: ${settings.postmortemThreshold}`,
		'custom_fields:',
		...settings.fields.map((field) => `  - name: "${field.name}"\n    type: ${field.type}`)
	];
	return lines.join('\n') + '\n';
}

export function applyImport(decision: string): boolean {
	return decision === 'replace' || decision === 'skip';
}
