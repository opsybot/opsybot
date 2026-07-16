import type { Severity } from '$lib/dashboard';
import {
	canMoveTo,
	type EntryType,
	type FollowUp,
	type Incident,
	type IncidentStage,
	type PostmortemState,
	type RelationKind,
	type StatusPageUpdate,
	type TimelineEntry
} from '$lib/incidents';
import { scenario } from './fixtures';

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

const iso = (offsetMs: number) => new Date(Date.now() + offsetMs).toISOString();

function seed(): { incidents: Incident[]; followUps: FollowUp[] } {
	const incidents: Incident[] = [
		{
			id: 'INC-2481',
			name: 'Checkout degraded — EU',
			severity: 'SEV1',
			status: 'investigating',
			lead: 'Maya Chen',
			comms: 'Priya Nair',
			team: 'payments',
			services: ['payments-api', 'checkout-web'],
			declaredAt: iso(-38 * MINUTE),
			nextUpdateAt: iso(6 * MINUTE),
			resolvedAt: null,
			mine: true,
			summary:
				'Checkout error rate in EU spiked to 12% at 09:12 UTC, ten minutes after deploy 84c2 landed on payments-api. Failover to us-east-1 replicas started 09:31; error rate is down to 3.1% and falling. EU deploys are held. Root cause suspected in the new payment-authorization retry logic.',
			customFields: [
				{ label: 'Customer impact', value: 'EU checkout · ~12% of sessions' },
				{ label: 'Detected by', value: 'synthetic monitoring' },
				{ label: 'Deploy under suspicion', value: '84c2', mono: true }
			],
			related: [
				{ relation: 'caused by', id: 'INC-2478', name: 'Bad deploy 84c2 rolled back' }
			],
			alerts: [
				{
					id: 'al-6',
					title: 'Synthetic checkout failing from us-east-1',
					severity: 'CRIT',
					tone: 'critical',
					status: 'acked'
				},
				{
					id: 'grp-1',
					title: 'payments-api p99 above 800 ms',
					severity: 'HIGH',
					tone: 'high',
					status: 'open'
				}
			],
			timeline: [
				{
					id: 't1',
					type: 'status',
					at: iso(-38 * MINUTE),
					actor: 'Maya Chen',
					text: 'Incident declared as SEV1 · lead Maya Chen'
				},
				{
					id: 't2',
					type: 'status',
					at: iso(-37 * MINUTE),
					actor: 'Maya Chen',
					text: 'Status → investigating'
				},
				{
					id: 't3',
					type: 'observation',
					at: iso(-34 * MINUTE),
					actor: 'Opsybot',
					ai: true,
					text: 'Correlated: deploy 84c2 to payments-api landed 09:02 UTC, 10 min before the first alert. Suggested starting point.'
				},
				{
					id: 't4',
					type: 'communication',
					at: iso(-30 * MINUTE),
					actor: 'Priya Nair',
					text: 'Posted to status page: elevated checkout errors in EU, investigating. Next update 15 min.'
				},
				{
					id: 't5',
					type: 'action',
					at: iso(-21 * MINUTE),
					actor: 'Marcus Lee',
					text: 'Initiated failover of payments-api to us-east-1 replicas',
					edited: true
				},
				{
					id: 't6',
					type: 'decision',
					at: iso(-17 * MINUTE),
					actor: 'Maya Chen',
					text: 'Holding all EU deploys until error rate is back under 0.5%'
				},
				{
					id: 't7',
					type: 'observation',
					at: iso(-8 * MINUTE),
					actor: 'Marcus Lee',
					text: 'Checkout error rate dropping: 12% → 3.1% over the last 5 min'
				}
			],
			statusPage: {
				domain: 'status.acme.dev',
				stage: 'identified',
				title: 'Elevated checkout errors in Europe',
				updates: [
					{
						at: iso(-14 * MINUTE),
						stage: 'Identified',
						text: 'The cause is identified and a fix is rolling out. EU checkout errors are decreasing.'
					},
					{
						at: iso(-30 * MINUTE),
						stage: 'Investigating',
						text: 'We are investigating elevated error rates on checkout in Europe.'
					}
				]
			},
			postmortem: 'not-started'
		},
		{
			id: 'INC-2480',
			name: 'Elevated 5xx on search',
			severity: 'SEV2',
			status: 'identified',
			lead: 'Marcus Lee',
			comms: 'Maya Chen',
			team: 'platform',
			services: ['gateway'],
			declaredAt: iso(-62 * MINUTE),
			nextUpdateAt: iso(22 * MINUTE),
			resolvedAt: null,
			mine: false,
			summary: '',
			customFields: [],
			related: [],
			alerts: [],
			timeline: [
				{
					id: 'u1',
					type: 'status',
					at: iso(-62 * MINUTE),
					actor: 'Marcus Lee',
					text: 'Incident declared as SEV2 · lead Marcus Lee'
				}
			],
			statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
			postmortem: 'not-started'
		},
		{
			id: 'INC-2479',
			name: 'Delayed webhook deliveries',
			severity: 'SEV3',
			status: 'monitoring',
			lead: 'Maya Chen',
			comms: 'Dev Patel',
			team: 'payments',
			services: ['events-worker'],
			declaredAt: iso(-3 * HOUR - 40 * MINUTE),
			nextUpdateAt: iso(-4 * MINUTE),
			resolvedAt: null,
			mine: true,
			summary: '',
			customFields: [],
			related: [],
			alerts: [],
			timeline: [
				{
					id: 'v1',
					type: 'status',
					at: iso(-3 * HOUR - 40 * MINUTE),
					actor: 'Maya Chen',
					text: 'Incident declared as SEV3 · lead Maya Chen'
				}
			],
			statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
			postmortem: 'not-started'
		},
		{
			id: 'INC-2478',
			name: 'Bad deploy 84c2 rolled back',
			severity: 'SEV2',
			status: 'resolved',
			lead: 'Priya Nair',
			comms: 'Maya Chen',
			team: 'platform',
			services: ['gateway'],
			declaredAt: iso(-3 * DAY),
			nextUpdateAt: null,
			resolvedAt: iso(-3 * DAY + HOUR + 47 * MINUTE),
			mine: false,
			summary: '',
			customFields: [
				{ label: 'Peak error rate', value: '4.2%' },
				{ label: 'Affected surface', value: 'public APIs' },
				{ label: 'Data loss', value: 'none' }
			],
			related: [],
			alerts: [],
			timeline: [
				{
					id: 'w1',
					type: 'status',
					at: iso(-3 * DAY),
					actor: 'Priya Nair',
					text: 'Incident declared as SEV2 · lead Priya Nair'
				},
				{
					id: 'w2',
					type: 'observation',
					at: iso(-3 * DAY + 2 * MINUTE),
					actor: 'Opsybot',
					text: 'Gateway error rate crossed 4.2% — up from a 0.1% baseline',
					ai: true
				},
				{
					id: 'w3',
					type: 'observation',
					at: iso(-3 * DAY + 17 * MINUTE),
					actor: 'Marcus Lee',
					text: 'Errors started within a minute of deploy 84c2 reaching all gateway instances'
				},
				{
					id: 'w4',
					type: 'decision',
					at: iso(-3 * DAY + 29 * MINUTE),
					actor: 'Priya Nair',
					text: 'Rolling back 84c2 rather than rolling forward — the regression is in routing'
				},
				{
					id: 'w5',
					type: 'action',
					at: iso(-3 * DAY + 60 * MINUTE),
					actor: 'Marcus Lee',
					text: 'Rollback requested — waiting on the manual approval gate'
				},
				{
					id: 'w6',
					type: 'action',
					at: iso(-3 * DAY + 82 * MINUTE),
					actor: 'Marcus Lee',
					text: 'Rollback approved and applied to every gateway instance'
				},
				{
					id: 'w7',
					type: 'communication',
					at: iso(-3 * DAY + 90 * MINUTE),
					actor: 'Maya Chen',
					text: 'Status page updated — errors subsiding, watching for 30 minutes'
				},
				{
					id: 'w8',
					type: 'status',
					at: iso(-3 * DAY + HOUR + 47 * MINUTE),
					actor: 'Priya Nair',
					text: 'Resolved — error rate back to baseline for 30 minutes'
				}
			],
			statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
			postmortem: 'published'
		},
		{
			id: 'INC-2475',
			name: 'Elevated 5xx on search',
			severity: 'SEV2',
			status: 'resolved',
			lead: 'Dev Patel',
			comms: 'Priya Nair',
			team: 'platform',
			services: ['gateway'],
			declaredAt: iso(-9 * DAY),
			nextUpdateAt: null,
			resolvedAt: iso(-9 * DAY + 40 * MINUTE),
			mine: false,
			summary: '',
			customFields: [],
			related: [],
			alerts: [],
			timeline: [
				{
					id: 'y1',
					type: 'status',
					at: iso(-9 * DAY),
					actor: 'Dev Patel',
					text: 'Incident declared as SEV2 · lead Dev Patel'
				},
				{
					id: 'y2',
					type: 'observation',
					at: iso(-9 * DAY + 11 * MINUTE),
					actor: 'Sana Ito',
					text: 'Search nodes were evicted by the scheduler and never rescheduled'
				},
				{
					id: 'y3',
					type: 'action',
					at: iso(-9 * DAY + 32 * MINUTE),
					actor: 'Dev Patel',
					text: 'Search cluster restarted, error rate fell immediately'
				},
				{
					id: 'y4',
					type: 'status',
					at: iso(-9 * DAY + 40 * MINUTE),
					actor: 'Dev Patel',
					text: 'Resolved — search responses back to baseline'
				}
			],
			statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
			postmortem: 'not-started'
		},
		{
			id: 'INC-2468',
			name: 'Database failover stalled',
			severity: 'SEV1',
			status: 'resolved',
			lead: 'Maya Chen',
			comms: 'Priya Nair',
			team: 'platform',
			services: ['database'],
			declaredAt: iso(-14 * DAY),
			nextUpdateAt: null,
			resolvedAt: iso(-14 * DAY + 3 * HOUR + 12 * MINUTE),
			mine: false,
			summary: '',
			customFields: [],
			related: [],
			alerts: [],
			timeline: [
				{
					id: 'z1',
					type: 'status',
					at: iso(-14 * DAY),
					actor: 'Maya Chen',
					text: 'Incident declared as SEV1 · lead Maya Chen'
				},
				{
					id: 'z2',
					type: 'status',
					at: iso(-14 * DAY + 3 * HOUR + 12 * MINUTE),
					actor: 'Maya Chen',
					text: 'Resolved — primary promoted by hand'
				}
			],
			statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
			postmortem: 'published'
		},
		{
			id: 'INC-2461',
			name: 'Cert expiry took down internal tools',
			severity: 'SEV3',
			status: 'resolved',
			lead: 'Sana Ito',
			comms: 'Dev Patel',
			team: 'platform',
			services: ['edge'],
			declaredAt: iso(-23 * DAY),
			nextUpdateAt: null,
			resolvedAt: iso(-23 * DAY + 55 * MINUTE),
			mine: false,
			summary: '',
			customFields: [],
			related: [],
			alerts: [],
			timeline: [
				{
					id: 'v1',
					type: 'status',
					at: iso(-23 * DAY),
					actor: 'Sana Ito',
					text: 'Incident declared as SEV3 · lead Sana Ito'
				},
				{
					id: 'v2',
					type: 'status',
					at: iso(-23 * DAY + 55 * MINUTE),
					actor: 'Sana Ito',
					text: 'Resolved — certificate renewed and pushed to the edge'
				}
			],
			statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
			postmortem: 'published'
		},
		{
			id: 'INC-2455',
			name: 'Search indexing job silently failed',
			severity: 'SEV2',
			status: 'resolved',
			lead: 'Marcus Lee',
			comms: 'Maya Chen',
			team: 'platform',
			services: ['events-worker'],
			declaredAt: iso(-31 * DAY),
			nextUpdateAt: null,
			resolvedAt: iso(-31 * DAY + 4 * HOUR),
			mine: false,
			summary: '',
			customFields: [],
			related: [],
			alerts: [],
			timeline: [
				{
					id: 'u1',
					type: 'status',
					at: iso(-31 * DAY),
					actor: 'Marcus Lee',
					text: 'Incident declared as SEV2 · lead Marcus Lee'
				},
				{
					id: 'u2',
					type: 'status',
					at: iso(-31 * DAY + 4 * HOUR),
					actor: 'Marcus Lee',
					text: 'Resolved — indexing backfilled and alerting added'
				}
			],
			statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
			postmortem: 'published'
		},
		{
			id: 'INC-2472',
			name: 'Rate limiter misconfigured',
			severity: 'SEV3',
			status: 'resolved',
			lead: 'Dev Patel',
			comms: 'Sana Ito',
			team: 'frontend',
			services: ['edge'],
			declaredAt: iso(-5 * DAY),
			nextUpdateAt: null,
			resolvedAt: iso(-5 * DAY + 3 * HOUR),
			mine: false,
			summary: '',
			customFields: [],
			related: [],
			alerts: [],
			timeline: [
				{
					id: 'x1',
					type: 'status',
					at: iso(-5 * DAY),
					actor: 'Dev Patel',
					text: 'Incident declared as SEV3 · lead Dev Patel'
				}
			],
			statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
			postmortem: 'draft'
		}
	];

	const followUps: FollowUp[] = [
		{
			id: 'f1',
			incidentId: 'INC-2481',
			title: 'Add canary stage to payments-api deploys',
			owner: 'Marcus Lee',
			dueAt: iso(5 * DAY),
			done: false
		},
		{
			id: 'f2',
			incidentId: 'INC-2481',
			title: 'Alert on checkout error rate, not just p99',
			owner: 'Maya Chen',
			dueAt: iso(2 * DAY),
			done: false
		},
		{
			id: 'f3',
			incidentId: 'INC-2478',
			title: 'Automate rollback on 5xx spike',
			owner: 'Priya Nair',
			dueAt: iso(-4 * DAY),
			done: false
		},
		{
			id: 'f4',
			incidentId: 'INC-2478',
			title: 'Document gateway failover runbook',
			owner: 'Dev Patel',
			dueAt: iso(-2 * DAY),
			done: false
		},
		{
			id: 'f5',
			incidentId: 'INC-2472',
			title: 'Add rate-limit config validation in CI',
			owner: 'Dev Patel',
			dueAt: iso(-7 * DAY),
			done: true
		}
	];

	return { incidents, followUps };
}

let store = seed();
let empty = !['active', 'degraded'].includes(scenario());
let nextId = 2482;

function id(prefix: string): string {
	return prefix + Math.random().toString(36).slice(2, 8);
}

function find(incidentId: string): Incident | undefined {
	return store.incidents.find((incident) => incident.id === incidentId);
}

function record(incident: Incident, type: EntryType, text: string, actor = 'Maya Chen') {
	incident.timeline.push({ id: id('t'), type, at: new Date().toISOString(), actor, text });
}

export function listIncidents(): Incident[] {
	return empty ? [] : store.incidents;
}

export function getIncident(incidentId: string): Incident | undefined {
	return empty ? undefined : find(incidentId);
}

export function listFollowUps(): FollowUp[] {
	return empty ? [] : store.followUps;
}

export function declareIncident(input: {
	name: string;
	severity: Severity;
	services: string[];
	lead: string;
	alerts: string[];
}): Incident {
	const incident: Incident = {
		id: `INC-${nextId++}`,
		name: input.name,
		severity: input.severity,
		status: 'declared',
		lead: input.lead,
		comms: input.lead,
		team: 'payments',
		services: input.services,
		declaredAt: new Date().toISOString(),
		nextUpdateAt: new Date(Date.now() + 30 * MINUTE).toISOString(),
		resolvedAt: null,
		mine: true,
		summary: '',
		customFields: [],
		related: [],
		alerts: [],
		timeline: [],
		statusPage: { domain: 'status.acme.dev', stage: 'none', title: '', updates: [] },
		postmortem: 'not-started'
	};

	record(
		incident,
		'status',
		`Incident declared as ${input.severity} · lead ${input.lead}`,
		input.lead
	);

	empty = false;
	store.incidents.unshift(incident);
	return incident;
}

export function rename(incidentId: string, name: string) {
	const incident = find(incidentId);
	if (!incident || !name.trim() || name === incident.name) return;
	incident.name = name.trim();
	record(incident, 'status', `Renamed to “${incident.name}”`);
}

export function changeSeverity(incidentId: string, severity: Severity) {
	const incident = find(incidentId);
	if (!incident || incident.severity === severity) return;
	incident.severity = severity;
	record(incident, 'status', `Severity changed to ${severity}`);
}

export function assignRole(incidentId: string, role: 'lead' | 'comms', person: string) {
	const incident = find(incidentId);
	if (!incident) return;
	incident[role] = person;
	record(
		incident,
		'status',
		role === 'lead' ? `Lead handed over to ${person}` : `Comms role assigned to ${person}`
	);
}

export function moveStatus(incidentId: string, status: IncidentStage) {
	const incident = find(incidentId);
	if (!incident || !canMoveTo(incident.status, status)) return;
	incident.status = status;
	record(incident, 'status', `Status → ${status}`);
}

export function postUpdate(incidentId: string) {
	const incident = find(incidentId);
	if (!incident) return;
	incident.nextUpdateAt = new Date(Date.now() + 15 * MINUTE).toISOString();
	record(
		incident,
		'communication',
		'Status update posted: fix rolling out, error rate falling. Next update in 15 min.'
	);
}

// The incident owns its status-page history; the status pages section writes through here
export function publishStatusUpdate(
	incidentId: string,
	stage: StatusPageUpdate['stage'],
	text: string,
	options: { domain?: string; title?: string } = {}
): boolean {
	const incident = find(incidentId);
	if (!incident) return false;

	if (options.domain) incident.statusPage.domain = options.domain;
	if (options.title && !incident.statusPage.title) incident.statusPage.title = options.title;
	incident.statusPage.stage = stage.toLowerCase();
	incident.statusPage.updates = [
		{ at: new Date().toISOString(), stage, text },
		...incident.statusPage.updates
	];

	record(incident, 'communication', `Status page updated — ${stage.toLowerCase()}`);
	return true;
}

export function resolveIncident(
	incidentId: string,
	summary: string,
	alsoAlerts: boolean,
	schedulePostmortem: boolean
) {
	const incident = find(incidentId);
	if (!incident) return;

	incident.status = 'resolved';
	incident.resolvedAt = new Date().toISOString();
	incident.nextUpdateAt = null;
	record(incident, 'status', `Resolved — ${summary}`);

	if (alsoAlerts && incident.alerts.length) {
		incident.alerts = incident.alerts.map((alert) => ({ ...alert, status: 'resolved' as const }));
		record(incident, 'status', `${incident.alerts.length} linked alerts resolved with the incident`);
	}

	if (schedulePostmortem) {
		incident.postmortem = 'not-started';
		record(
			incident,
			'status',
			`Postmortem scheduled — draft due in 3 working days, assigned to ${incident.lead}`
		);
	}
}

export function reopenIncident(incidentId: string) {
	const incident = find(incidentId);
	if (!incident) return;
	incident.status = 'monitoring';
	incident.resolvedAt = null;
	incident.nextUpdateAt = new Date(Date.now() + 15 * MINUTE).toISOString();
	record(incident, 'status', 'Reopened — update reminders restarted');
}

export function addEntry(
	incidentId: string,
	type: EntryType,
	text: string,
	options?: { at?: string; retro?: boolean }
) {
	const incident = find(incidentId);
	if (!incident || !text.trim()) return;

	const entry: TimelineEntry = {
		id: id('t'),
		type,
		at: options?.at ?? new Date().toISOString(),
		actor: 'Maya Chen',
		text: text.trim(),
		retro: options?.retro
	};

	incident.timeline.push(entry);
	incident.timeline.sort((a, b) => Date.parse(a.at) - Date.parse(b.at));
}

export function linkIncident(incidentId: string, relation: RelationKind, otherId: string) {
	const incident = find(incidentId);
	const other = find(otherId);
	if (!incident || !other || incident.id === other.id) return;
	if (incident.related.some((entry) => entry.id === otherId)) return;

	incident.related.push({ relation, id: other.id, name: other.name });
	record(incident, 'status', `Linked ${otherId} as ${relation}`);
}

export function addFollowUp(incidentId: string, title: string, owner: string, dueAt: string) {
	if (!title.trim()) return;
	store.followUps.push({
		id: id('f'),
		incidentId,
		title: title.trim(),
		owner,
		dueAt,
		done: false
	});
}

export function toggleFollowUp(followUpId: string) {
	const followUp = store.followUps.find((entry) => entry.id === followUpId);
	if (followUp) followUp.done = !followUp.done;
}

export function setPostmortem(incidentId: string, state: PostmortemState) {
	const incident = find(incidentId);
	if (incident) incident.postmortem = state;
}
