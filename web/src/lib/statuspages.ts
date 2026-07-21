import type { Tone } from '$lib/dashboard';

export type Visibility = 'public' | 'token' | 'auth';

export const VISIBILITIES: { value: Visibility; label: string; description: string }[] = [
	{ value: 'public', label: 'Public', description: 'Anyone with the URL can view.' },
	{ value: 'token', label: 'Private: link token', description: 'Only people with the tokenized link.' },
	{ value: 'auth', label: 'Private: authenticated', description: 'Viewers sign in with a workspace account.' }
];

export type ComponentState = 'operational' | 'degraded' | 'partial' | 'major';

export const COMPONENT_STATES: { value: Exclude<ComponentState, 'operational'>; label: string }[] = [
	{ value: 'degraded', label: 'Degraded performance' },
	{ value: 'partial', label: 'Partial outage' },
	{ value: 'major', label: 'Major outage' }
];

export const COMPONENT_STATE_LABEL: Record<ComponentState, string> = {
	operational: 'Operational',
	degraded: 'Degraded performance',
	partial: 'Partial outage',
	major: 'Major outage'
};

export const COMPONENT_STATE_TONE: Record<ComponentState, Tone> = {
	operational: 'success',
	degraded: 'warning',
	partial: 'warning',
	major: 'critical'
};

const SEVERITY_ORDER: ComponentState[] = ['major', 'partial', 'degraded', 'operational'];

export type Component = {
	id: string;
	name: string;
	group: string;
	services: string[];
	state: ComponentState;
};

export type StatusPage = {
	id: string;
	name: string;
	description: string;
	pageTitle: string;
	visibility: Visibility;
	domain: string;
	domainVerified: boolean;
	certRenews: string;
	accent: string;
	utcDefault: boolean;
	showUptime: boolean;
	allowIndexing: boolean;
	token: string;
	published: boolean;
	components: Component[];
	subscribers: { email: number; feed: number; webhook: number };
};

export const ACCENTS = [
	{ id: 'mint', color: '#00E5AC' },
	{ id: 'blue', color: '#4DA3FF' },
	{ id: 'violet', color: '#9A7DFF' },
	{ id: 'amber', color: '#F5B23D' }
];

export const NOTICES = ['24 h before', '48 h before', '48 h + 1 h before', '72 h before'];

export const NEXT_UPDATE = ['15 min', '30 min', '60 min'];

export const PUBLISH_STAGES = ['investigating', 'identified', 'monitoring', 'resolved'] as const;
export type PublishStage = (typeof PUBLISH_STAGES)[number];

export const STAGE_CAPITAL: Record<PublishStage, 'Investigating' | 'Identified' | 'Monitoring' | 'Resolved'> = {
	investigating: 'Investigating',
	identified: 'Identified',
	monitoring: 'Monitoring',
	resolved: 'Resolved'
};

export const STAGE_TONE: Record<PublishStage, Tone> = {
	investigating: 'critical',
	identified: 'warning',
	monitoring: 'info',
	resolved: 'success'
};

export const TEMPLATES: Record<PublishStage, string> = {
	investigating:
		'We are investigating [symptom customers see]. [Scope: regions/products affected.] Next update by [time] UTC.',
	identified:
		'The cause is identified and a fix is in progress. [Current customer impact.] Next update by [time] UTC.',
	monitoring:
		'A fix is deployed and metrics are recovering. We are monitoring before calling this resolved. Next update by [time] UTC.',
	resolved:
		'This incident is resolved. [One line: what happened, duration.] A postmortem will follow if impact warranted one.'
};

export type Overall = 'operational' | 'degraded' | 'outage';

export const OVERALL: Record<Overall, { label: string; tone: Tone }> = {
	operational: { label: 'all systems operational', tone: 'success' },
	degraded: { label: 'degraded performance', tone: 'warning' },
	outage: { label: 'major outage', tone: 'critical' }
};

export function overallOf(page: StatusPage): Overall {
	const worst = page.components
		.map((component) => component.state)
		.sort((a, b) => SEVERITY_ORDER.indexOf(a) - SEVERITY_ORDER.indexOf(b))[0];

	if (worst === 'major') return 'outage';
	if (worst === 'degraded' || worst === 'partial') return 'degraded';
	return 'operational';
}

export function subscriberTotal(page: StatusPage): number {
	const { email, feed, webhook } = page.subscribers;
	return email + feed + webhook;
}
