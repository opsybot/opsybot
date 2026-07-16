import type { Incident, PostmortemState, TimelineEntry } from '$lib/incidents';
import type { Factor, Postmortem, SectionId } from '$lib/postmortems';
import { postmortemId } from '$lib/postmortems';
import { getIncident, listIncidents, setPostmortem } from './incidents';

const id = (prefix: string) => prefix + Math.random().toString(36).slice(2, 8);

const clock = (entry: TimelineEntry) => `${entry.at.slice(11, 16)} UTC`;

const of = (incident: Incident, type: TimelineEntry['type']) =>
	incident.timeline.filter((entry) => entry.type === type);

function factor(label: string, text: string, fromTimeline: string[] = []): Factor {
	return { id: id('fa-'), label, text, fromTimeline };
}

function refs(incidentId: string, entryIds: string[]): string[] {
	const incident = getIncident(incidentId);
	if (!incident) return [];

	return incident.timeline
		.filter((entry) => entryIds.includes(entry.id))
		.map((entry) => `${clock(entry)} ${entry.type}`);
}

function seed(): Postmortem[] {
	return [
		{
			id: 'PM-2478',
			incidentId: 'INC-2478',
			author: 'Priya Nair',
			summary:
				'Deploy 84c2 introduced a routing regression in the gateway. Error rates rose to 4.2% across public APIs for 1 h 47 m until the deploy was rolled back. No data was lost.',
			impact:
				'Roughly 8% of API consumers saw intermittent 502s during the impact window. 61 customer requests reached the support inbox. Internal tooling was unaffected.',
			wentWell:
				'Alerting fired within 90 seconds of the error-rate rise. The on-call acknowledged in under a minute and the incident channel had full context from the start.',
			improve:
				'Rollbacks should be one command. Canary deploys would have caught this with 1% of the blast radius.',
			factors: [
				factor(
					'no canary stage',
					'The deploy pipeline has no canary stage — 84c2 reached 100% of gateway instances in one step.',
					refs('INC-2478', ['w3', 'w4'])
				),
				factor(
					'slow rollback tooling',
					'Rollback tooling required manual approval, adding 22 minutes to recovery.',
					refs('INC-2478', ['w5', 'w6'])
				)
			],
			announce: true,
			publicLink: false,
			publishedAt: new Date(Date.now() - 3 * 86_400_000 + 5 * 3_600_000).toISOString(),
			history: [
				{
					id: 'h-1',
					at: new Date(Date.now() - 3 * 86_400_000 + 4 * 3_600_000).toISOString(),
					by: 'Priya Nair',
					what: 'Draft created from the timeline'
				},
				{
					id: 'h-2',
					at: new Date(Date.now() - 3 * 86_400_000 + 4.5 * 3_600_000).toISOString(),
					by: 'Priya Nair',
					what: 'Edited: impact, contributing factors'
				},
				{
					id: 'h-3',
					at: new Date(Date.now() - 3 * 86_400_000 + 5 * 3_600_000).toISOString(),
					by: 'Marcus Lee',
					what: 'Published'
				}
			]
		},
		{
			id: 'PM-2468',
			incidentId: 'INC-2468',
			author: 'Maya Chen',
			summary:
				'An automated failover stalled halfway. The primary was promoted by hand 3 h 12 m after the first alert.',
			impact: 'Writes failed for the duration. Reads served stale data from the replica.',
			wentWell: 'The replica held, so nothing was lost when the primary was finally promoted.',
			improve: 'The failover runbook described a topology we no longer run.',
			factors: [
				factor(
					'stale or missing runbook',
					'The failover runbook described a topology that had been replaced two quarters earlier.'
				),
				factor(
					'no canary stage',
					'The failover path is exercised only in production, so a broken step is found during an incident.'
				)
			],
			announce: true,
			publicLink: false,
			publishedAt: new Date(Date.now() - 13 * 86_400_000).toISOString(),
			history: []
		},
		{
			id: 'PM-2461',
			incidentId: 'INC-2461',
			author: 'Sana Ito',
			summary:
				'An expired certificate took the internal tooling domain offline for 55 minutes. Customer traffic was unaffected.',
			impact: 'Internal dashboards and the admin console were unreachable. No customer impact.',
			wentWell: 'The fix was well understood; renewal took minutes once the cause was clear.',
			improve: 'Nothing was watching the expiry date, so the first signal was the outage itself.',
			factors: [
				factor(
					'no expiry alerting',
					'Certificate expiry was tracked in a spreadsheet, and nothing alerted on it.'
				),
				factor(
					'config change without validation',
					'The renewal was applied by hand with no check that the edge had picked it up.'
				)
			],
			announce: false,
			publicLink: false,
			publishedAt: new Date(Date.now() - 22 * 86_400_000).toISOString(),
			history: []
		},
		{
			id: 'PM-2455',
			incidentId: 'INC-2455',
			author: 'Marcus Lee',
			summary:
				'The nightly search-indexing job failed for six days without anyone noticing. Search results went stale.',
			impact: 'Search returned results up to six days out of date. No errors were raised.',
			wentWell: 'The backfill was clean once the failure was found.',
			improve: 'A job that fails silently is a job nobody is running.',
			factors: [
				factor(
					'config change without validation',
					'A credential rotation changed the job’s environment, and nothing validated that it still ran.'
				),
				factor(
					'stale or missing runbook',
					'There was no runbook for the indexing job, so the first responder had to read the source.'
				)
			],
			announce: false,
			publicLink: false,
			publishedAt: new Date(Date.now() - 30 * 86_400_000).toISOString(),
			history: []
		}
	];
}

const store = seed();

export function publishedOn(row: { postmortem: Postmortem; incident: Incident }): string {
	return row.postmortem.publishedAt ?? row.incident.resolvedAt ?? row.incident.declaredAt;
}

export function listPublished(): { postmortem: Postmortem; incident: Incident }[] {
	return listIncidents()
		.filter((incident) => incident.postmortem === 'published')
		// openPostmortem back-fills a draft for incidents published without a record
		.map((incident) => openPostmortem(incident.id))
		.filter((row): row is { postmortem: Postmortem; incident: Incident } => !!row)
		.sort((a, b) => Date.parse(publishedOn(b)) - Date.parse(publishedOn(a)));
}

function getByIncident(incidentId: string): Postmortem {
	return store.find((postmortem) => postmortem.incidentId === incidentId) as Postmortem;
}

export function getPostmortem(
	pmId: string
): { postmortem: Postmortem; incident: Incident } | undefined {
	const postmortem = store.find((entry) => entry.id === pmId);
	if (!postmortem) return;

	const incident = getIncident(postmortem.incidentId);
	return incident ? { postmortem, incident } : undefined;
}

function record(postmortem: Postmortem, what: string, by = 'Maya Chen') {
	postmortem.history.push({ id: id('h-'), at: new Date().toISOString(), by, what });
}

export function openPostmortem(
	incidentId: string,
	by = 'Maya Chen'
): { postmortem: Postmortem; incident: Incident } | undefined {
	const incident = getIncident(incidentId);
	if (!incident) return;

	const existing = getByIncident(incidentId);
	if (existing) return { postmortem: existing, incident };

	const postmortem: Postmortem = {
		id: postmortemId(incidentId),
		incidentId,
		author: by,
		summary: draft(incident, 'summary'),
		impact: draft(incident, 'impact'),
		wentWell: draft(incident, 'wentWell'),
		improve: draft(incident, 'improve'),
		factors: proposeFactors(incident),
		announce: true,
		publicLink: false,
		publishedAt: null,
		history: []
	};

	record(postmortem, 'Draft created from the timeline', by);
	store.push(postmortem);

	if (incident.postmortem === 'not-started') setPostmortem(incidentId, 'draft');

	return { postmortem, incident };
}

function sentence(text: string): string {
	return /[.!?"')\]]$/.test(text.trim()) ? text.trim() : `${text.trim()}.`;
}

function resolvedClock(incident: Incident): string | null {
	return incident.resolvedAt ? `${incident.resolvedAt.slice(11, 16)} UTC` : null;
}

function windowOf(incident: Incident): string {
	const first = incident.timeline[0];
	const end = resolvedClock(incident);

	if (!first) return 'the impact window';
	if (end && end !== clock(first)) return `${clock(first)} – ${end}`;
	return `from ${clock(first)}`;
}

export function draft(incident: Incident, section: SectionId): string {
	const window = windowOf(incident);

	if (section === 'summary') {
		const cause = of(incident, 'observation')[0] ?? of(incident, 'decision')[0];
		const fix = of(incident, 'action').at(-1);
		const end = resolvedClock(incident);

		return [
			sentence(
				`${incident.name}, declared ${incident.severity} on ${incident.services.join(', ') || 'the service'}`
			),
			cause ? sentence(cause.text) : null,
			fix ? sentence(`${fix.text} at ${clock(fix)}`) : null,
			end ? `Resolved at ${end}, ${window}.` : 'Still open.'
		]
			.filter(Boolean)
			.join(' ');
	}

	if (section === 'impact') {
		const measured = incident.customFields
			.filter((field) => /\d/.test(field.value))
			.map((field) => `${field.label.toLowerCase()} ${field.value}`)
			.join(', ');

		return [
			`${incident.services.join(', ') || 'The affected service'} was degraded ${window}.`,
			measured ? `Recorded at the time: ${measured}.` : 'No impact numbers were recorded.',
			'Fill in who this reached and how they felt it.'
		].join(' ');
	}

	if (section === 'wentWell') {
		const acted = of(incident, 'action').length;
		const decided = of(incident, 'decision').length;

		return [
			`The timeline records ${acted} ${acted === 1 ? 'action' : 'actions'} and ${decided} ${decided === 1 ? 'decision' : 'decisions'}, so the response was written down as it happened.`,
			'Say what made that possible, so it survives the next reorganisation.'
		].join(' ');
	}

	const slow = of(incident, 'action')[0];

	return [
		slow ? `The first action was taken at ${clock(slow)}.` : null,
		'Say what would have made this shorter, in terms of the system rather than the people.'
	]
		.filter(Boolean)
		.join(' ');
}

function proposeFactors(incident: Incident): Factor[] {
	return [...of(incident, 'observation'), ...of(incident, 'decision')].slice(0, 2).map((entry) =>
		factor('', entry.text, [`${clock(entry)} ${entry.type}`])
	);
}

export function writeSection(pmId: string, section: SectionId, text: string, by = 'Maya Chen') {
	const found = getPostmortem(pmId);
	if (!found) return;

	found.postmortem[section] = text;
	record(found.postmortem, `Edited: ${section === 'wentWell' ? 'what went well' : section === 'improve' ? 'what could be improved' : section}`, by);
}

export function addFactor(pmId: string, by = 'Maya Chen') {
	const found = getPostmortem(pmId);
	if (!found) return;

	found.postmortem.factors.push(factor('', '', []));
	record(found.postmortem, 'Added a contributing factor', by);
}

export function writeFactor(pmId: string, factorId: string, patch: Partial<Factor>) {
	const found = getPostmortem(pmId);
	if (!found) return;

	found.postmortem.factors = found.postmortem.factors.map((entry) =>
		entry.id === factorId ? { ...entry, ...patch } : entry
	);
}

export function removeFactor(pmId: string, factorId: string, by = 'Maya Chen') {
	const found = getPostmortem(pmId);
	if (!found) return;

	found.postmortem.factors = found.postmortem.factors.filter((entry) => entry.id !== factorId);
	record(found.postmortem, 'Removed a contributing factor', by);
}

export function setOptions(pmId: string, options: { announce?: boolean; publicLink?: boolean }) {
	const found = getPostmortem(pmId);
	if (!found) return;

	Object.assign(found.postmortem, options);
}

export function advance(
	incidentId: string,
	to: PostmortemState,
	options: { reviewer?: string; by?: string } = {}
) {
	const by = options.by ?? 'Maya Chen';

	const found = to === 'not-started' ? getPostmortem(postmortemId(incidentId)) : openPostmortem(incidentId, by);
	if (!found) return;

	if (to === found.incident.postmortem) return;

	setPostmortem(incidentId, to);

	if (to === 'in-review') {
		record(found.postmortem, options.reviewer ? `Sent to ${options.reviewer} for review` : 'Sent for review', by);
	} else if (to === 'published') {
		found.postmortem.publishedAt = new Date().toISOString();
		record(found.postmortem, 'Published', by);
	}
}

export function requestReview(pmId: string, reviewer: string, by = 'Maya Chen') {
	const found = getPostmortem(pmId);
	if (found) advance(found.incident.id, 'in-review', { reviewer, by });
}

export function publish(pmId: string, by = 'Maya Chen') {
	const found = getPostmortem(pmId);
	if (found) advance(found.incident.id, 'published', { by });
}
