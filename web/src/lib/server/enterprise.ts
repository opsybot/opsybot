import type { AuditData, Perm, RolesData, ScimData, ScimEvent, SecurityPolicy } from '$lib/enterprise';
import { scenario } from './fixtures';

const EVENTS: ScimEvent[] = [
	{ at: '2026-07-11 09:00:12 UTC', kind: 'sync', text: 'Full sync: 42 users, 3 groups, no drift', tone: 'success' },
	{ at: '2026-07-10 16:41:03 UTC', kind: 'create', text: 'Provisioned jordan@acme.dev from group eng-all → viewer', tone: 'success' },
	{ at: '2026-07-09 11:22:47 UTC', kind: 'deprovision', text: 'Deprovisioned tom@acme.dev (left IdP group)', tone: 'warning', wizard: '3 references reassigned to Sana Ito' },
	{ at: '2026-07-08 09:00:09 UTC', kind: 'update', text: 'Updated 4 users. Name/email changes from IdP', tone: 'neutral' }
];

const ROLES = ['Stakeholder', 'Responder', 'Team admin', 'Billing admin', 'sec-auditor (custom)'];
const PERMS: Perm[] = [
	{ perm: 'View incidents and timelines', grants: [1, 1, 1, 0, 1] },
	{ perm: 'Ack, resolve, and declare', grants: [0, 1, 1, 0, 0] },
	{ perm: 'Edit schedules and policies for their team', grants: [0, 0, 1, 0, 0] },
	{ perm: 'Manage members and teams', grants: [0, 0, 1, 0, 0] },
	{ perm: 'Read the audit log', grants: [0, 0, 1, 0, 1] },
	{ perm: 'Manage billing and licenses', grants: [0, 0, 0, 1, 0] },
	{ perm: 'Configure SSO, SCIM, security policies', grants: [0, 0, 0, 0, 0] }
];

const store = {
	licensed: scenario() !== 'empty',
	scim: {
		endpoint: 'https://opsy.bot/scim/v2/acme',
		token: 'scim_9f27c3a1e8b44d17aa02',
		lastSync: '2026-07-11 09:00 UTC · 42 users · 3 groups · IdP: sso.acme.dev',
		events: EVENTS
	} as ScimData,
	audit: {
		savedFilters: [
			{ name: 'Auth failures: 30 d', q: 'action:auth.login.failed' },
			{ name: 'Config changes by API keys', q: 'actor:*-provider action:config.*' },
			{ name: 'Role changes', q: 'action:member.role.*' }
		],
		streamEndpoint: 'https://siem.acme.dev/ingest/opsybot',
		format: 'JSON lines',
		retention: '7 years'
	} as AuditData,
	security: {
		sessionLifetime: '14 days',
		ipAllowlist: '10.0.0.0/8\n84.121.9.0/24  # office',
		enforce2fa: true,
		ssoEnforced: false
	} as SecurityPolicy
};

function hex(length: number): string {
	let out = '';
	while (out.length < length) out += Math.floor(Math.random() * 16).toString(16);
	return out.slice(0, length);
}

export function isLicensed(): boolean {
	return store.licensed;
}

export function getScim(): ScimData {
	return store.scim;
}

export function getRoles(): RolesData {
	return { roles: ROLES, perms: PERMS };
}

export function getAudit(): AuditData {
	return store.audit;
}

export function getSecurity(): SecurityPolicy {
	return store.security;
}

export function rotateScimToken(): string {
	store.scim = { ...store.scim, token: `scim_${hex(20)}` };
	return store.scim.token;
}

export function saveSecurityPolicy(policy: SecurityPolicy): void {
	store.security = policy;
}
