import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
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
import { apiClient } from './api';

type Schemas = components['schemas'];

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

function mapMember(dto: Schemas['Member']): Member {
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
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/members', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map(mapMember);
}

export async function inviteMember(
	cookies: Cookies,
	workspace: string,
	email: string,
	role: Role
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/members/invites', {
		params: { path: { workspaceId: workspace } },
		body: { email, role }
	});
	return error ? { error: error.detail ?? 'Could not invite that person.' } : {};
}

export async function changeRole(
	cookies: Cookies,
	workspace: string,
	userId: string,
	role: Role
): Promise<boolean> {
	const { error } = await apiClient(cookies).PUT('/workspaces/{workspaceId}/members/{userId}/role', {
		params: { path: { workspaceId: workspace, userId } },
		body: { role }
	});
	return !error;
}

export async function deactivateMember(
	cookies: Cookies,
	workspace: string,
	userId: string,
	replacements: Record<string, string>
): Promise<boolean> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/members/{userId}/deactivate',
		{ params: { path: { workspaceId: workspace, userId } }, body: { replacements } }
	);
	return !error;
}

export async function reactivateMember(
	cookies: Cookies,
	workspace: string,
	userId: string
): Promise<boolean> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/members/{userId}/reactivate',
		{ params: { path: { workspaceId: workspace, userId } } }
	);
	return !error;
}

async function memberNameIndex(cookies: Cookies, workspace: string) {
	const members = await listMembers(cookies, workspace);
	const byId = new Map(members.map((member) => [member.id, member.name]));
	const byName = new Map(members.map((member) => [member.name, member.id]));
	return { byId, byName };
}

function mapTeam(dto: Schemas['Team'], byId: Map<string, string>): Team {
	return {
		id: dto.slug,
		name: dto.name,
		members: dto.memberIds.map((id) => byId.get(id) ?? id),
		archived: dto.archived,
		schedules: [],
		policies: [],
		services: []
	};
}

export async function listTeams(
	cookies: Cookies,
	workspace: string,
	includeArchived = false
): Promise<Team[]> {
	const [{ data }, { byId }] = await Promise.all([
		apiClient(cookies).GET('/workspaces/{workspaceId}/teams', {
			params: { path: { workspaceId: workspace }, query: { includeArchived: includeArchived || undefined } }
		}),
		memberNameIndex(cookies, workspace)
	]);
	return (data?.items ?? []).map((team) => mapTeam(team, byId));
}

export async function archiveTeam(cookies: Cookies, workspace: string, slug: string): Promise<boolean> {
	const { error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/teams/{teamSlug}/archive', {
		params: { path: { workspaceId: workspace, teamSlug: slug } }
	});
	return !error;
}

export async function unarchiveTeam(cookies: Cookies, workspace: string, slug: string): Promise<boolean> {
	const { error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/teams/{teamSlug}/unarchive', {
		params: { path: { workspaceId: workspace, teamSlug: slug } }
	});
	return !error;
}

export async function getTeam(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<Team | undefined> {
	const [{ data }, { byId }] = await Promise.all([
		apiClient(cookies).GET('/workspaces/{workspaceId}/teams/{teamSlug}', {
			params: { path: { workspaceId: workspace, teamSlug: slug } }
		}),
		memberNameIndex(cookies, workspace)
	]);
	return data ? mapTeam(data, byId) : undefined;
}

export async function createTeam(
	cookies: Cookies,
	workspace: string,
	name: string,
	memberNames: string[]
): Promise<{ id?: string; error?: string }> {
	const { byName } = await memberNameIndex(cookies, workspace);
	const memberIds = memberNames.map((n) => byName.get(n)).filter((id): id is string => Boolean(id));
	const { data, error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/teams', {
		params: { path: { workspaceId: workspace } },
		body: { name, memberIds }
	});
	if (error || !data) return { error: error?.detail ?? 'Could not create the team.' };
	return { id: data.slug };
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
	const { error } = await apiClient(cookies).PATCH('/workspaces/{workspaceId}/teams/{teamSlug}', {
		params: { path: { workspaceId: workspace, teamSlug: slug } },
		body: { name, memberIds }
	});
	return !error;
}

function mapKey(dto: Schemas['ApiKey']): ApiKey {
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
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/keys', {
		params: { path: { workspaceId: workspace } }
	});
	return {
		personal: (data?.personal ?? []).map(mapKey),
		workspace: (data?.workspace ?? []).map(mapKey)
	};
}

export async function createKey(
	cookies: Cookies,
	workspace: string,
	name: string,
	scopes: string[],
	kind: KeyKind
): Promise<{ secret?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/keys', {
		params: { path: { workspaceId: workspace } },
		body: { name, kind, scopes: scopes as Schemas['Scope'][] }
	});
	if (error || !data) return { error: error?.detail ?? 'Could not create the key.' };
	return { secret: data.secret };
}

export async function revokeKey(cookies: Cookies, workspace: string, id: string): Promise<boolean> {
	const { error } = await apiClient(cookies).DELETE('/workspaces/{workspaceId}/keys/{keyId}', {
		params: { path: { workspaceId: workspace, keyId: id } }
	});
	return !error;
}

export type AuditQuery = { q?: string; actor?: string; action?: string; cursor?: string };

export async function listAudit(
	cookies: Cookies,
	workspace: string,
	query: AuditQuery = {}
): Promise<{ entries: AuditEntry[]; nextCursor: string }> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/audit', {
		params: {
			path: { workspaceId: workspace },
			query: {
				limit: 50,
				q: query.q || undefined,
				actor: query.actor || undefined,
				action: query.action || undefined,
				cursor: query.cursor || undefined
			}
		}
	});
	return {
		entries: (data?.items ?? []).map((entry) => ({
			id: entry.id,
			at: entry.at.replace('T', ' ').slice(0, 19) + ' UTC',
			actor: entry.actor || 'system',
			action: entry.action,
			target: entry.target,
			ip: entry.ip || '–'
		})),
		nextCursor: data?.nextCursor ?? ''
	};
}

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
	sso: {
		mode: 'oidc',
		issuer: '',
		clientId: '',
		hasClientSecret: false,
		clientSecret: '',
		clearClientSecret: false,
		samlMetadataUrl: '',
		scopes: '',
		allowedEmailDomains: '',
		enabled: false,
		required: false,
		jitProvisioning: false
	},
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
		{ path: 'policies/frontend-daytime', note: 'references user tom@acme.dev, now deactivated. Pick a replacement or skip.' }
	],
	skipped: [{ path: 'statuspages/status.acme.dev', note: 'identical to current config' }]
};

let settings = DEFAULT_SETTINGS();

function splitList(raw: string): string[] {
	return raw
		.split(/[\s,]+/)
		.map((value) => value.trim())
		.filter(Boolean);
}

export async function getSettings(cookies: Cookies, workspace: string): Promise<WorkspaceSettings> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/sso', {
		params: { path: { workspaceId: workspace } }
	});
	if (data) {
		settings.sso = {
			mode: data.mode === 'saml' ? 'saml' : 'oidc',
			issuer: data.issuer,
			clientId: data.clientId,
			hasClientSecret: data.hasClientSecret,
			clientSecret: '',
			clearClientSecret: false,
			samlMetadataUrl: data.samlMetadataUrl,
			scopes: (data.scopes ?? []).join(', '),
			allowedEmailDomains: (data.allowedEmailDomains ?? []).join('\n'),
			enabled: data.enabled,
			required: data.required,
			jitProvisioning: data.jitProvisioning
		};
	}
	return settings;
}

export async function saveSettings(
	cookies: Cookies,
	workspace: string,
	next: WorkspaceSettings
): Promise<void> {
	settings = { ...next, fields: settings.fields };
	const s = next.sso;
	const body: Schemas['SsoConfigRequest'] = {
		mode: s.mode,
		issuer: s.issuer,
		clientId: s.clientId,
		samlMetadataUrl: s.samlMetadataUrl,
		scopes: splitList(s.scopes),
		allowedEmailDomains: splitList(s.allowedEmailDomains),
		enabled: s.enabled,
		required: s.required,
		jitProvisioning: s.jitProvisioning,
		clearClientSecret: s.clearClientSecret,
		...(s.clientSecret ? { clientSecret: s.clientSecret } : {})
	};
	await apiClient(cookies).PUT('/workspaces/{workspaceId}/sso', {
		params: { path: { workspaceId: workspace } },
		body
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
