import type { Tone } from '$lib/dashboard';

let idSeq = 0;
export function uid(prefix = 'x'): string {
	idSeq += 1;
	return `${prefix}${idSeq.toString(36)}-${Math.random().toString(36).slice(2, 7)}`;
}

export function slugify(value: string): string {
	return value
		.toLowerCase()
		.trim()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '')
		.slice(0, 40);
}

export type Role = 'admin' | 'member';
export const ROLES: Role[] = ['admin', 'member'];

export type MemberReference = { id: string; kind: string; icon: string; label: string; detail: string };

export type Member = {
	id: string;
	name: string;
	email: string;
	role: Role;
	twoFactor: boolean;
	auth: string;
	lastActive: string;
	references: MemberReference[];
	deactivated: boolean;
};

export function isRole(value: string): value is Role {
	return (ROLES as string[]).includes(value);
}

export function twoFactorBadge(on: boolean): { tone: Tone; label: string } {
	return on ? { tone: 'success', label: 'on' } : { tone: 'warning', label: 'off' };
}

export function parseInvite(form: FormData): { email: string; role: Role } | { error: string } {
	const email = String(form.get('email') ?? '').trim().slice(0, 120);
	const role = String(form.get('role') ?? '');
	if (!/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email)) return { error: 'Enter a valid email address.' };
	if (!isRole(role)) return { error: 'Pick a role from the list.' };
	return { email, role };
}

export type Team = {
	id: string;
	name: string;
	members: string[];
	archived: boolean;
	schedules: string[];
	policies: string[];
	services: string[];
};

export function teamSummary(team: Team): string {
	const schedules = team.schedules.length;
	const services = team.services.length;
	return `${team.members.length} members · ${schedules} ${schedules === 1 ? 'schedule' : 'schedules'} · ${services} services`;
}

export function parseTeam(form: FormData, roster: string[]): { name: string; members: string[] } | { error: string } {
	const name = slugify(String(form.get('name') ?? ''));
	if (!name) return { error: 'Give the team a name.' };
	let members: unknown;
	try {
		members = JSON.parse(String(form.get('members') ?? '[]'));
	} catch {
		members = [];
	}
	const allowed = new Set(roster);
	const list = Array.isArray(members)
		? members.filter((m): m is string => typeof m === 'string' && allowed.has(m)).slice(0, 50)
		: [];
	return { name, members: [...new Set(list)] };
}

export type Scope = string;
export const SCOPES: Scope[] = [
	'incidents:read',
	'incidents:write',
	'alerts:read',
	'alerts:write',
	'schedules:write',
	'policies:write',
	'config:write',
	'audit:read'
];
const SCOPE_SET = new Set(SCOPES);

export type KeyKind = 'personal' | 'workspace';
export function isKeyKind(value: string): value is KeyKind {
	return value === 'personal' || value === 'workspace';
}

export type ApiKey = { id: string; name: string; scopes: Scope[]; last: string; created: string };

export function parseKeyDraft(
	form: FormData
): { name: string; scopes: Scope[]; kind: KeyKind } | { error: string } {
	const name = String(form.get('name') ?? '').replace(/\s+/g, ' ').trim().slice(0, 60);
	const kind = String(form.get('kind') ?? '');
	let raw: unknown;
	try {
		raw = JSON.parse(String(form.get('scopes') ?? '[]'));
	} catch {
		raw = [];
	}
	const scopes = Array.isArray(raw) ? raw.filter((s): s is string => typeof s === 'string' && SCOPE_SET.has(s)) : [];
	if (!name) return { error: 'Give the key a name.' };
	if (!isKeyKind(kind)) return { error: 'Unknown key kind.' };
	if (scopes.length === 0) return { error: 'Pick at least one scope.' };
	return { name, scopes: [...new Set(scopes)], kind };
}

export type AuditEntry = { id: string; at: string; actor: string; action: string; target: string; ip: string };

export function isFailure(action: string): boolean {
	return action.includes('failed');
}

export type Severity = { id: string; def: string };
export const SEVERITY_TONE: Record<string, Tone> = { SEV1: 'critical', SEV2: 'high', SEV3: 'warning', SEV4: 'info' };

export type CustomField = { id: string; name: string; type: string; options?: string };
export const FIELD_TYPES = ['text', 'select', 'multi-select', 'number'];

export const THRESHOLD_OPTIONS = ['SEV1', 'SEV2', 'SEV3'];
export const CADENCE_OPTIONS = ['15 min', '30 min', '1 h', '2 h', 'none'];
export const TIMEZONE_OPTIONS = ['UTC', 'Europe/Berlin', 'America/New_York', 'Asia/Tokyo'];
export const RETENTION_OPTIONS = {
	alerts: ['90 days', '1 year', '2 years', 'forever'],
	incidents: ['1 year', '2 years', 'forever'],
	audit: ['1 year', '2 years', 'forever']
};

export type WorkspaceSettings = {
	name: string;
	timezone: string;
	severities: Severity[];
	postmortemThreshold: string;
	fields: CustomField[];
	cadence: Record<string, string>;
	sso: SsoSettings;
	retention: { alerts: string; incidents: string; audit: string };
};

export type SsoSettings = {
	mode: 'oidc' | 'saml';
	issuer: string;
	clientId: string;
	hasClientSecret: boolean;
	clientSecret: string;
	clearClientSecret: boolean;
	samlMetadataUrl: string;
	scopes: string;
	allowedEmailDomains: string;
	enabled: boolean;
	required: boolean;
	jitProvisioning: boolean;
};

const inList = (value: unknown, options: readonly string[], fallback: string) =>
	typeof value === 'string' && options.includes(value) ? value : fallback;

export function parseSettings(raw: string, current: WorkspaceSettings): WorkspaceSettings | { error: string } {
	let data: unknown;
	try {
		data = JSON.parse(raw);
	} catch {
		return { error: 'Could not read the settings.' };
	}
	if (!data || typeof data !== 'object') return { error: 'Could not read the settings.' };
	const obj = data as Record<string, unknown>;
	const severities = Array.isArray(obj.severities)
		? current.severities.map((sev, index) => {
				const submitted = obj.severities as unknown[];
				const entry = submitted[index] && typeof submitted[index] === 'object' ? (submitted[index] as Record<string, unknown>) : {};
				return { id: sev.id, def: typeof entry.def === 'string' ? entry.def.slice(0, 200) : sev.def };
			})
		: current.severities;
	const cadence: Record<string, string> = {};
	for (const sev of Object.keys(current.cadence)) {
		const submitted = obj.cadence && typeof obj.cadence === 'object' ? (obj.cadence as Record<string, unknown>) : {};
		cadence[sev] = inList(submitted[sev], CADENCE_OPTIONS, current.cadence[sev]);
	}
	const sso = obj.sso && typeof obj.sso === 'object' ? (obj.sso as Record<string, unknown>) : {};
	const retention = obj.retention && typeof obj.retention === 'object' ? (obj.retention as Record<string, unknown>) : {};
	return {
		name: typeof obj.name === 'string' && obj.name.trim() ? obj.name.trim().slice(0, 80) : current.name,
		timezone: inList(obj.timezone, TIMEZONE_OPTIONS, current.timezone),
		severities,
		postmortemThreshold: inList(obj.postmortemThreshold, THRESHOLD_OPTIONS, current.postmortemThreshold),
		fields: current.fields,
		cadence,
		sso: {
			mode: sso.mode === 'saml' ? 'saml' : 'oidc',
			issuer: typeof sso.issuer === 'string' ? sso.issuer.slice(0, 200) : current.sso.issuer,
			clientId: typeof sso.clientId === 'string' ? sso.clientId.slice(0, 120) : current.sso.clientId,
			hasClientSecret: current.sso.hasClientSecret,
			clientSecret: typeof sso.clientSecret === 'string' ? sso.clientSecret.slice(0, 400) : '',
			clearClientSecret: sso.clearClientSecret === true,
			samlMetadataUrl: typeof sso.samlMetadataUrl === 'string' ? sso.samlMetadataUrl.slice(0, 300) : current.sso.samlMetadataUrl,
			scopes: typeof sso.scopes === 'string' ? sso.scopes.slice(0, 300) : current.sso.scopes,
			allowedEmailDomains:
				typeof sso.allowedEmailDomains === 'string' ? sso.allowedEmailDomains.slice(0, 500) : current.sso.allowedEmailDomains,
			enabled: sso.enabled === true,
			required: sso.required === true,
			jitProvisioning: sso.jitProvisioning === true
		},
		retention: {
			alerts: inList(retention.alerts, RETENTION_OPTIONS.alerts, current.retention.alerts),
			incidents: inList(retention.incidents, RETENTION_OPTIONS.incidents, current.retention.incidents),
			audit: inList(retention.audit, RETENTION_OPTIONS.audit, current.retention.audit)
		}
	};
}

export type DiffKind = 'created' | 'changed' | 'decision' | 'skipped';
export type DiffItem = { path: string; note: string };
export type ConfigDiff = Record<DiffKind, DiffItem[]>;

export const DIFF_GROUPS: { kind: DiffKind; label: string; tone: Tone; icon: string }[] = [
	{ kind: 'created', label: 'Created', tone: 'success', icon: 'plus' },
	{ kind: 'changed', label: 'Changed', tone: 'info', icon: 'pencil' },
	{ kind: 'decision', label: 'Needs decision', tone: 'warning', icon: 'triangle-alert' },
	{ kind: 'skipped', label: 'Skipped', tone: 'neutral', icon: 'minus' }
];

export const IMPORT_DECISIONS = [
	{ value: 'replace', label: 'Replace with Sana Ito' },
	{ value: 'skip', label: 'Skip this resource' }
];
