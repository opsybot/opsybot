import type { Tone } from '$lib/dashboard';

export type RangeKey = '30d' | '90d' | '1y';

export const RANGE_OPTIONS: { value: RangeKey; label: string }[] = [
	{ value: '30d', label: 'Last 30 days' },
	{ value: '90d', label: 'Last quarter' },
	{ value: '1y', label: 'Last year' }
];

export const TEAMS = ['payments', 'platform', 'frontend'];
export const SERVICES = ['payments-api', 'gateway', 'database'];
export const SEVERITIES = ['SEV1', 'SEV2', 'SEV3'];

// K-anonymity floor: below this cohort size, on-call load is withheld
export const COHORT_FLOOR = 5;

export type Filters = {
	team: string;
	service: string;
	severity: string;
	range: RangeKey;
};

function isRange(value: string | null): value is RangeKey {
	return value === '30d' || value === '90d' || value === '1y';
}

const oneOf = (value: string | null, allowed: string[]): string =>
	value && allowed.includes(value) ? value : '';

export function parseFilters(url: URL): Filters {
	const range = url.searchParams.get('range');
	return {
		team: oneOf(url.searchParams.get('team'), TEAMS),
		service: oneOf(url.searchParams.get('service'), SERVICES),
		severity: oneOf(url.searchParams.get('severity'), SEVERITIES),
		range: isRange(range) ? range : '30d'
	};
}

export function filterQuery(filters: Filters, overrides: Partial<Filters> = {}): string {
	const merged = { ...filters, ...overrides };
	const params = new URLSearchParams();
	if (merged.team) params.set('team', merged.team);
	if (merged.service) params.set('service', merged.service);
	if (merged.severity) params.set('severity', merged.severity);
	if (merged.range !== '30d') params.set('range', merged.range);
	const query = params.toString();
	return query ? `?${query}` : '';
}

export function scopeLabel(filters: Filters): string {
	const parts: string[] = [];
	if (filters.team) parts.push(filters.team);
	if (filters.service) parts.push(filters.service);
	if (filters.severity) parts.push(filters.severity);
	parts.push(RANGE_OPTIONS.find((option) => option.value === filters.range)!.label.toLowerCase());
	return parts.join(' · ');
}

export type Definition = {
	key: string;
	term: string;
	blurb: string;
	definition: string;
};

export const DEFINITIONS: Definition[] = [
	{
		key: 'MTTA',
		term: 'MTTA — mean time to acknowledge',
		blurb: 'From page sent to a human acknowledging. Measures how fast someone picks up.',
		definition:
			'Median duration from the first page sent to a human acknowledging it. Excludes auto-resolved alerts that never paged.'
	},
	{
		key: 'MTTR',
		term: 'MTTR — mean time to resolve',
		blurb: 'From declare to resolved. The headline recovery number.',
		definition:
			'Median from incident declared to resolved. Reopened incidents count the full span including the reopen.'
	},
	{
		key: 'TTID',
		term: 'Time to identified',
		blurb: 'From investigating to identified — how long root-causing takes.',
		definition:
			'From the investigating stage to identified. Only incidents that reached identified are included.'
	},
	{
		key: 'escalation',
		term: 'Escalation depth',
		blurb: 'How many escalation steps fired before someone acknowledged.',
		definition:
			'How many escalation steps fired before acknowledgement. Depth 1 means the first person got it.'
	},
	{
		key: 'night',
		term: 'Night pages',
		blurb: 'Pages delivered overnight in the responder’s own timezone.',
		definition: 'Pages delivered 22:00–07:00 in the responder’s own timezone.'
	},
	{
		key: 'hours',
		term: 'On-call hours',
		blurb: 'Scheduled on-call time, whether or not anything paged.',
		definition:
			'Scheduled on-call time, whether or not anything paged. Overrides are attributed to whoever actually held the shift.'
	}
];

const DEFINITION_BY_KEY = new Map(DEFINITIONS.map((entry) => [entry.key, entry]));

export function definitionBlurb(key: string): string {
	return DEFINITION_BY_KEY.get(key)?.blurb ?? '';
}

export type MetricCard = {
	key: string;
	label: string;
	value: string;
	delta: string;
	good: boolean;
};

export type StageRow = { label: string; value: string; pct: number };

export type Overview = {
	metrics: MetricCard[];
	mttrTrend: number[];
	stages: StageRow[];
};

export type AlertStat = { key: string; value: string; note: string };

export type AlertAnalytics = {
	volume: number[];
	stats: AlertStat[];
};

export type LoadRow = {
	name: string;
	team: string;
	hours: number;
	pages: number;
	night: number;
	weekend: number;
};

export type OnCallLoad = {
	rows: LoadRow[];
	note: string;
};

export type FollowupStat = { key: string; value: string; tone: Extract<Tone, 'success' | 'warning' | 'neutral'> };
export type TeamCompletion = { team: string; pct: number };

export type FollowupCompletion = {
	stats: FollowupStat[];
	byTeam: TeamCompletion[];
};

export function barTone(pct: number): 'brand' | 'warning' {
	return pct < 70 ? 'warning' : 'brand';
}

export function toCsv(rows: (string | number)[][]): string {
	const cell = (value: string | number) => {
		const text = String(value);
		return /[",\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
	};
	return rows.map((row) => row.map(cell).join(',')).join('\r\n');
}
