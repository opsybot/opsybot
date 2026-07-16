import type { Severity, Tone } from '$lib/dashboard';
import type { Incident, PostmortemState } from '$lib/incidents';

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

export type SectionId = 'summary' | 'impact' | 'wentWell' | 'improve';

export type Factor = {
	id: string;
	label: string;
	text: string;
	fromTimeline: string[];
};

export type Edit = { id: string; at: string; by: string; what: string };

export type Postmortem = {
	id: string;
	incidentId: string;
	author: string;
	summary: string;
	impact: string;
	wentWell: string;
	improve: string;
	factors: Factor[];
	announce: boolean;
	publicLink: boolean;
	publishedAt: string | null;
	history: Edit[];
};

export const REVIEWERS = ['Marcus Lee', 'Maya Chen', 'Dev Patel', 'Sana Ito'];

export const STATE_TONE: Record<PostmortemState, Tone | 'brand'> = {
	'not-started': 'neutral',
	draft: 'neutral',
	'in-review': 'info',
	published: 'success'
};

export function stateLabel(state: PostmortemState): string {
	return state === 'in-review' ? 'in review' : state === 'not-started' ? 'not started' : state;
}

export function owesPostmortem(incident: { severity: Severity }): boolean {
	return incident.severity === 'SEV1' || incident.severity === 'SEV2';
}

const WORKING_DAY_LIMIT = 3;

// Weekends and the resolve day itself do not count
function workingDaysSince(iso: string, now: number): number {
	let days = 0;

	for (let at = Date.parse(iso) + DAY; at <= now; at += DAY) {
		const day = new Date(at).getUTCDay();
		if (day !== 0 && day !== 6) days++;
	}

	return days;
}

export type Waiting = {
	incidentId: string;
	title: string;
	severity: Severity;
	state: string;
	overdue: boolean;
	resolved: string;
};

export function waitingOn(incident: Incident, now: number): Waiting | null {
	if (!owesPostmortem(incident)) return null;
	if (incident.postmortem === 'published') return null;

	if (!incident.resolvedAt) {
		return {
			incidentId: incident.id,
			title: incident.name,
			severity: incident.severity,
			state: 'waiting on resolve',
			overdue: false,
			resolved: 'not yet resolved'
		};
	}

	const late = workingDaysSince(incident.resolvedAt, now) - WORKING_DAY_LIMIT;
	if (late <= 0) return null;

	return {
		incidentId: incident.id,
		title: incident.name,
		severity: incident.severity,
		state: `${late} ${late === 1 ? 'day' : 'days'} overdue`,
		overdue: true,
		resolved: `resolved ${incident.resolvedAt.slice(0, 10)}`
	};
}

export type Pattern = { label: string; count: number; postmortems: string[] };

export function patterns(published: Postmortem[]): Pattern[] {
	const seen = new Map<string, string[]>();

	for (const postmortem of published) {
		for (const label of new Set(postmortem.factors.map((factor) => factor.label.trim()))) {
			if (!label) continue;
			seen.set(label, [...(seen.get(label) ?? []), postmortem.id]);
		}
	}

	return [...seen.entries()]
		.filter(([, ids]) => ids.length > 1)
		.map(([label, ids]) => ({ label, count: ids.length, postmortems: ids }))
		.sort((a, b) => b.count - a.count || a.label.localeCompare(b.label));
}

export function formatDuration(incident: Incident): string {
	if (!incident.resolvedAt) return 'still open';

	// Round to whole minutes before splitting so 1 h 59 m 40 s cannot read 1 h 60 m
	const total = Math.round((Date.parse(incident.resolvedAt) - Date.parse(incident.declaredAt)) / MINUTE);
	const hours = Math.floor(total / 60);
	const minutes = total % 60;

	return hours ? `${hours} h ${minutes} m` : `${minutes} m`;
}

export function impactWindow(incident: Incident): string {
	const utc = (iso: string) => `${iso.slice(0, 10)} ${iso.slice(11, 16)}`;

	return incident.resolvedAt
		? `${utc(incident.declaredAt)} – ${incident.resolvedAt.slice(11, 16)} UTC`
		: `${utc(incident.declaredAt)} UTC – open`;
}

export function responders(incident: Incident): string {
	const others = incident.timeline
		.map((entry) => entry.actor)
		.filter((actor) => actor !== 'Opsybot' && actor !== incident.lead);

	return [`${incident.lead} (lead)`, ...new Set(others)].join(', ');
}

export function facts(incident: Incident): { label: string; value: string; mono?: boolean }[] {
	return [
		{ label: 'Severity', value: incident.severity },
		{ label: 'Duration', value: formatDuration(incident) },
		{ label: 'Impact window', value: impactWindow(incident), mono: true },
		{ label: 'Responders', value: responders(incident) },
		{ label: 'Services', value: incident.services.join(', ') || 'none recorded' }
	];
}

export function postmortemId(incidentId: string): string {
	return incidentId.replace('INC-', 'PM-');
}
