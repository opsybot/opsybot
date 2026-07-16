import type { Tone } from '$lib/dashboard';

export type EntTab = 'scim' | 'roles' | 'audit' | 'security';

export const ENT_TABS: { id: EntTab; label: string; icon: string }[] = [
	{ id: 'scim', label: 'SCIM provisioning', icon: 'users' },
	{ id: 'roles', label: 'Fine-grained roles', icon: 'shield-check' },
	{ id: 'audit', label: 'Advanced audit', icon: 'history' },
	{ id: 'security', label: 'Security policies', icon: 'lock' }
];

export const ENT_PITCH: Record<EntTab, string> = {
	scim: 'Provision and deprovision from your identity provider automatically — leavers lose access the moment IdP says so, with references safely reassigned.',
	roles: 'Stakeholder, team admin, billing admin, and custom roles — least-privilege access described in plain language.',
	audit: 'Saved filters, exports, SIEM streaming, and multi-year retention on top of the core audit log.',
	security: 'Session lifetimes, admin IP allowlists, enforced 2FA, and SSO-only login with break-glass recovery.'
};

export type ScimEventKind = 'sync' | 'create' | 'deprovision' | 'update';
export type ScimEvent = { at: string; kind: ScimEventKind; text: string; tone: Tone; wizard?: string };

export function scimEventColor(tone: Tone): string {
	return tone === 'warning' ? 'var(--warning)' : tone === 'success' ? 'var(--success)' : 'var(--text-tertiary)';
}

export type ScimData = {
	endpoint: string;
	token: string;
	lastSync: string;
	events: ScimEvent[];
};

export type Perm = { perm: string; grants: number[] };
export type RolesData = { roles: string[]; perms: Perm[] };

export type SavedFilter = { name: string; q: string };
export const FORMAT_OPTIONS = ['JSON lines', 'CEF (syslog)', 'OCSF'];
export const RETENTION_OPTIONS = ['1 year', '2 years', '7 years', 'forever'];

export type AuditData = {
	savedFilters: SavedFilter[];
	streamEndpoint: string;
	format: string;
	retention: string;
};

export const SESSION_OPTIONS = ['24 hours', '7 days', '14 days', '30 days'];

export type SecurityPolicy = {
	sessionLifetime: string;
	ipAllowlist: string;
	enforce2fa: boolean;
	ssoEnforced: boolean;
};

const inList = (value: unknown, options: readonly string[], fallback: string) =>
	typeof value === 'string' && options.includes(value) ? value : fallback;

export function parseSecurityPolicy(form: FormData, current: SecurityPolicy): SecurityPolicy {
	return {
		sessionLifetime: inList(form.get('sessionLifetime'), SESSION_OPTIONS, current.sessionLifetime),
		ipAllowlist: String(form.get('ipAllowlist') ?? '').replace(/\r\n/g, '\n').slice(0, 2000),
		enforce2fa: String(form.get('enforce2fa') ?? '') === 'true',
		ssoEnforced: String(form.get('ssoEnforced') ?? '') === 'true'
	};
}
