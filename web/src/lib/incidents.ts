import type { Severity, Tone } from '$lib/dashboard';

export type IncidentStage =
	| 'declared'
	| 'investigating'
	| 'identified'
	| 'monitoring'
	| 'resolved';

export const STAGES: IncidentStage[] = [
	'declared',
	'investigating',
	'identified',
	'monitoring',
	'resolved'
];

export const SEVERITIES: { id: Severity; tone: Tone; definition: string }[] = [
	{
		id: 'SEV1',
		tone: 'critical',
		definition: 'Customer-facing outage or data loss. All hands, page immediately.'
	},
	{
		id: 'SEV2',
		tone: 'high',
		definition: 'Major degradation for many customers. Page the on-call now.'
	},
	{
		id: 'SEV3',
		tone: 'warning',
		definition: 'Partial or contained impact. Fix during working hours.'
	},
	{ id: 'SEV4', tone: 'info', definition: 'Minor issue, no customer impact yet. Track it.' }
];

export const PEOPLE = ['Maya Chen', 'Priya Nair', 'Marcus Lee', 'Dev Patel', 'Sana Ito'];

export const SERVICES = [
	'payments-api',
	'gateway',
	'checkout-web',
	'database',
	'events-worker',
	'edge'
];

export const TEAMS = ['payments', 'platform', 'frontend'];

export type EntryType = 'status' | 'communication' | 'action' | 'observation' | 'decision';

export const ENTRY_TYPES: { id: EntryType; label: string }[] = [
	{ id: 'status', label: 'Status' },
	{ id: 'communication', label: 'Communication' },
	{ id: 'action', label: 'Action' },
	{ id: 'observation', label: 'Observation' },
	{ id: 'decision', label: 'Decision' }
];

export type TimelineEntry = {
	id: string;
	type: EntryType;
	at: string;
	actor: string;
	text: string;
	ai?: boolean;
	retro?: boolean;
	edited?: boolean;
};

export type LinkedAlert = {
	id: string;
	title: string;
	severity: string;
	tone: Tone;
	status: 'open' | 'acked' | 'resolved';
};

export type FollowUp = {
	id: string;
	incidentId: string;
	title: string;
	owner: string;
	dueAt: string;
	done: boolean;
};

export type RelationKind = 'caused by' | 'duplicates' | 'related to';

export type RelatedIncident = {
	relation: RelationKind;
	id: string;
	name: string;
	relationId?: string;
};

export type StatusPageUpdate = {
	at: string;
	stage: 'Investigating' | 'Identified' | 'Monitoring' | 'Resolved';
	text: string;
};

export type PostmortemState = 'not-started' | 'draft' | 'in-review' | 'published';

export type Incident = {
	id: string;
	ref?: string;
	name: string;
	severity: Severity;
	status: IncidentStage;
	lead: string;
	leadUserId?: string;
	comms: string;
	team: string;
	services: string[];
	declaredAt: string;
	nextUpdateAt: string | null;
	resolvedAt: string | null;
	mine: boolean;
	summary: string;
	customFields: { label: string; value: string; mono?: boolean }[];
	customFieldsRaw?: Record<string, string>;
	related: RelatedIncident[];
	alerts: LinkedAlert[];
	timeline: TimelineEntry[];
	statusPage: { domain: string; stage: string; title: string; updates: StatusPageUpdate[] };
	postmortem: PostmortemState;
};

export const RUNBOOKS = [
	{ service: 'payments-api', label: 'Checkout failures runbook' },
	{ service: 'payments-api', label: 'payments-api failover procedure' },
	{ service: 'checkout-web', label: 'Frontend error triage' }
];

export function isActive(incident: { status: IncidentStage }): boolean {
	return incident.status !== 'resolved';
}

export function canMoveTo(current: IncidentStage, target: IncidentStage): boolean {
	const from = STAGES.indexOf(current);
	const to = STAGES.indexOf(target);
	if (target === 'resolved') return from === STAGES.indexOf('monitoring');
	if (to === from + 1) return true;
	return to === from - 1 && target !== 'declared';
}
