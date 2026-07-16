import type {
	AlertAnalytics,
	Filters,
	FollowupCompletion,
	MetricCard,
	OnCallLoad,
	Overview,
	RangeKey,
	StageRow
} from '$lib/insights';
import { COHORT_FLOOR, DEFINITIONS, RANGE_OPTIONS } from '$lib/insights';
import { scenario } from './fixtures';

const available = scenario() !== 'empty';

export function insightsAvailable(): boolean {
	return available;
}

type Unit = 'sec' | 'min' | 'pct' | 'days' | 'count' | 'int';
type Kind = 'cost' | 'rate' | 'neutral';

type RawMetric = {
	key: string;
	label: string;
	value: number;
	unit: Unit;
	kind: Kind;
	delta: number;
	good: boolean;
};

const pad = (n: number) => String(n).padStart(2, '0');

function fmtSeconds(total: number): string {
	if (total < 60) return `${total}s`;
	const minutes = Math.floor(total / 60);
	const seconds = total % 60;
	if (minutes < 60) return seconds ? `${minutes}m ${pad(seconds)}s` : `${minutes}m`;
	const hours = Math.floor(minutes / 60);
	return `${hours}h ${pad(minutes % 60)}m`;
}

function fmtMinutes(total: number): string {
	if (total < 60) return `${total}m`;
	return `${Math.floor(total / 60)}h ${pad(total % 60)}m`;
}

function fmtValue(value: number, unit: Unit): string {
	switch (unit) {
		case 'sec':
			return fmtSeconds(Math.round(value));
		case 'min':
			return fmtMinutes(Math.round(value));
		case 'pct':
			return `${Math.round(value)}%`;
		case 'days': {
			const days = Math.round(value);
			return `${days} ${days === 1 ? 'day' : 'days'}`;
		}
		case 'int':
			return String(Math.round(value));
		case 'count':
			return Number.isInteger(value) ? String(value) : value.toFixed(1);
	}
}

const fmtDelta = (delta: number) => `${delta > 0 ? '+' : ''}${delta}%`;

const TEAM_WEIGHT: Record<string, number> = { payments: 1.15, platform: 0.95, frontend: 0.85 };
const SERVICE_WEIGHT: Record<string, number> = { 'payments-api': 1.2, gateway: 1.0, database: 1.1 };
const SEVERITY_WEIGHT: Record<string, number> = { SEV1: 1.45, SEV2: 1.0, SEV3: 0.6 };

function weight(filters: Filters): number {
	let factor = 1;
	if (filters.team) factor *= TEAM_WEIGHT[filters.team] ?? 1;
	if (filters.service) factor *= SERVICE_WEIGHT[filters.service] ?? 1;
	if (filters.severity) factor *= SEVERITY_WEIGHT[filters.severity] ?? 1;
	return factor;
}

function applyWeight(value: number, kind: Kind, factor: number): number {
	if (factor === 1 || kind === 'neutral') return value;
	if (kind === 'rate') return Math.max(0, Math.min(100, value / factor));
	return value * factor;
}

function card(metric: RawMetric, factor: number): MetricCard {
	return {
		key: metric.key,
		label: metric.label,
		value: fmtValue(applyWeight(metric.value, metric.kind, factor), metric.unit),
		delta: fmtDelta(metric.delta),
		good: metric.good
	};
}

const MTTR_TREND = [62, 58, 55, 71, 49, 52, 48, 44, 51, 47, 45, 48]; // minutes, last 12 weeks (fixed)
const ALERT_VOLUME = [120, 98, 140, 165, 132, 88, 76, 154, 142, 119, 103, 97, 128, 150]; // per day, last 14 days (fixed)

const OVERVIEW_METRICS: Record<RangeKey, RawMetric[]> = {
	'30d': [
		{ key: 'MTTA', label: 'Mean time to acknowledge', value: 134, unit: 'sec', kind: 'cost', delta: -18, good: true },
		{ key: 'MTTR', label: 'Mean time to resolve', value: 48, unit: 'min', kind: 'cost', delta: -9, good: true },
		{ key: 'TTID', label: 'Time to identified', value: 14, unit: 'min', kind: 'cost', delta: 6, good: false }
	],
	'90d': [
		{ key: 'MTTA', label: 'Mean time to acknowledge', value: 151, unit: 'sec', kind: 'cost', delta: -12, good: true },
		{ key: 'MTTR', label: 'Mean time to resolve', value: 53, unit: 'min', kind: 'cost', delta: -4, good: true },
		{ key: 'TTID', label: 'Time to identified', value: 15, unit: 'min', kind: 'cost', delta: 9, good: false }
	],
	'1y': [
		{ key: 'MTTA', label: 'Mean time to acknowledge', value: 182, unit: 'sec', kind: 'cost', delta: -22, good: true },
		{ key: 'MTTR', label: 'Mean time to resolve', value: 64, unit: 'min', kind: 'cost', delta: -15, good: true },
		{ key: 'TTID', label: 'Time to identified', value: 17, unit: 'min', kind: 'cost', delta: -3, good: true }
	]
};

// pct is an authored visual weight, not proportional to the value
const OVERVIEW_STAGES: Record<RangeKey, StageRow[]> = {
	'30d': [
		{ label: 'Declared → investigating', value: '1m 02s', pct: 8 },
		{ label: 'Investigating → identified', value: '14m', pct: 62 },
		{ label: 'Identified → monitoring', value: '19m', pct: 84 },
		{ label: 'Monitoring → resolved', value: '13m', pct: 58 }
	],
	'90d': [
		{ label: 'Declared → investigating', value: '1m 06s', pct: 9 },
		{ label: 'Investigating → identified', value: '15m', pct: 64 },
		{ label: 'Identified → monitoring', value: '20m', pct: 86 },
		{ label: 'Monitoring → resolved', value: '14m', pct: 60 }
	],
	'1y': [
		{ label: 'Declared → investigating', value: '1m 12s', pct: 10 },
		{ label: 'Investigating → identified', value: '17m', pct: 66 },
		{ label: 'Identified → monitoring', value: '22m', pct: 88 },
		{ label: 'Monitoring → resolved', value: '15m', pct: 62 }
	]
};

const ALERT_STATS: Record<RangeKey, RawMetric[]> = {
	'30d': [
		{ key: 'Ack rate', label: 'of pages acknowledged inside SLA', value: 96, unit: 'pct', kind: 'rate', delta: 0, good: true },
		{ key: 'Auto-resolved', label: 'closed before a human acted', value: 41, unit: 'pct', kind: 'neutral', delta: 0, good: true },
		{ key: 'Mean escalation depth', label: 'steps before someone acks', value: 1.4, unit: 'count', kind: 'cost', delta: 0, good: true }
	],
	'90d': [
		{ key: 'Ack rate', label: 'of pages acknowledged inside SLA', value: 94, unit: 'pct', kind: 'rate', delta: 0, good: true },
		{ key: 'Auto-resolved', label: 'closed before a human acted', value: 39, unit: 'pct', kind: 'neutral', delta: 0, good: true },
		{ key: 'Mean escalation depth', label: 'steps before someone acks', value: 1.5, unit: 'count', kind: 'cost', delta: 0, good: true }
	],
	'1y': [
		{ key: 'Ack rate', label: 'of pages acknowledged inside SLA', value: 92, unit: 'pct', kind: 'rate', delta: 0, good: true },
		{ key: 'Auto-resolved', label: 'closed before a human acted', value: 44, unit: 'pct', kind: 'neutral', delta: 0, good: true },
		{ key: 'Mean escalation depth', label: 'steps before someone acks', value: 1.6, unit: 'count', kind: 'cost', delta: 0, good: true }
	]
};

const FOLLOWUP_STATS: Record<RangeKey, (RawMetric & { tone: 'success' | 'warning' | 'neutral' })[]> = {
	'30d': [
		{ key: 'Completed on time', label: '', value: 71, unit: 'pct', kind: 'rate', delta: 0, good: true, tone: 'success' },
		{ key: 'Overdue', label: '', value: 5, unit: 'int', kind: 'cost', delta: 0, good: true, tone: 'warning' },
		{ key: 'Median time to close', label: '', value: 6, unit: 'days', kind: 'cost', delta: 0, good: true, tone: 'neutral' }
	],
	'90d': [
		{ key: 'Completed on time', label: '', value: 68, unit: 'pct', kind: 'rate', delta: 0, good: true, tone: 'success' },
		{ key: 'Overdue', label: '', value: 8, unit: 'int', kind: 'cost', delta: 0, good: true, tone: 'warning' },
		{ key: 'Median time to close', label: '', value: 7, unit: 'days', kind: 'cost', delta: 0, good: true, tone: 'neutral' }
	],
	'1y': [
		{ key: 'Completed on time', label: '', value: 74, unit: 'pct', kind: 'rate', delta: 0, good: true, tone: 'success' },
		{ key: 'Overdue', label: '', value: 3, unit: 'int', kind: 'cost', delta: 0, good: true, tone: 'warning' },
		{ key: 'Median time to close', label: '', value: 5, unit: 'days', kind: 'cost', delta: 0, good: true, tone: 'neutral' }
	]
};

const FOLLOWUP_BY_TEAM: { team: string; pct: number }[] = [
	{ team: 'payments', pct: 82 },
	{ team: 'platform', pct: 68 },
	{ team: 'frontend', pct: 74 }
];

type Person = { name: string; team: string; hours: number; pages: number; night: number; weekend: number };

const LOAD_PEOPLE: Person[] = [
	{ name: 'Priya Nair', team: 'payments', hours: 168, pages: 22, night: 6, weekend: 8 },
	{ name: 'Maya Chen', team: 'payments', hours: 168, pages: 18, night: 4, weekend: 8 },
	{ name: 'Marcus Lee', team: 'platform', hours: 120, pages: 14, night: 3, weekend: 4 },
	{ name: 'Dev Patel', team: 'platform', hours: 96, pages: 9, night: 2, weekend: 4 },
	{ name: 'Sana Ito', team: 'frontend', hours: 72, pages: 5, night: 1, weekend: 2 }
];

const RANGE_MULTIPLIER: Record<RangeKey, number> = { '30d': 1, '90d': 3, '1y': 12 };

const rangeLabel = (range: RangeKey) => RANGE_OPTIONS.find((option) => option.value === range)!.label;
const rangeWindow = (range: RangeKey) => ({ '30d': '30 days', '90d': 'the last quarter', '1y': 'the last year' }[range]);
const rangePrevious = (range: RangeKey) =>
	({ '30d': 'vs previous 30 days', '90d': 'vs previous quarter', '1y': 'vs previous year' }[range]);

export function getOverview(filters: Filters): Overview & { comparison: string } {
	const factor = weight(filters);
	return {
		metrics: OVERVIEW_METRICS[filters.range].map((metric) => card(metric, factor)),
		mttrTrend: MTTR_TREND,
		stages: OVERVIEW_STAGES[filters.range],
		comparison: rangePrevious(filters.range)
	};
}

export function getAlertAnalytics(filters: Filters): AlertAnalytics {
	const factor = weight(filters);
	return {
		volume: ALERT_VOLUME,
		stats: ALERT_STATS[filters.range].map((metric) => ({
			key: metric.key,
			value: fmtValue(applyWeight(metric.value, metric.kind, factor), metric.unit),
			note: metric.label
		}))
	};
}

export function getFollowupCompletion(filters: Filters): FollowupCompletion {
	const factor = weight(filters);
	const byTeam = filters.team ? FOLLOWUP_BY_TEAM.filter((row) => row.team === filters.team) : FOLLOWUP_BY_TEAM;
	return {
		stats: FOLLOWUP_STATS[filters.range].map((metric) => ({
			key: metric.key,
			value: fmtValue(applyWeight(metric.value, metric.kind, factor), metric.unit),
			tone: metric.tone
		})),
		byTeam
	};
}

// Rows are withheld below COHORT_FLOOR so individuals cannot be singled out
export function getOnCallLoad(filters: Filters): {
	rows: (OnCallLoad['rows'][number] & { heavy: boolean })[];
	note: string;
	header: string;
	footnote: string | null;
	withheld: boolean;
} {
	const multiplier = RANGE_MULTIPLIER[filters.range];
	const cohort = filters.team ? LOAD_PEOPLE.filter((person) => person.team === filters.team) : LOAD_PEOPLE;
	const header = `Load per person · ${rangeLabel(filters.range).toLowerCase()}`;
	const note = `Framed for load-balancing, not surveillance. Numbers aggregate over ${rangeWindow(
		filters.range
	)} with a floor of ${COHORT_FLOOR} people per view — never used to rank individuals.`;

	if (cohort.length < COHORT_FLOOR) {
		return { rows: [], note, header, footnote: null, withheld: true };
	}

	const rows = cohort.map((person) => ({
		name: person.name,
		team: person.team,
		hours: person.hours * multiplier,
		pages: person.pages * multiplier,
		night: person.night * multiplier,
		weekend: person.weekend * multiplier,
		// heavy checks the 30d base, not the scaled value, so the flag is stable across windows
		heavy: person.night > 5
	}));

	const heaviest = rows.reduce((worst, row) => (row.night > worst.night ? row : worst), rows[0]);
	const footnote = heaviest.heavy
		? `${heaviest.name} carried ${heaviest.night} night pages — worth rebalancing the ${heaviest.team} rotation before burnout, not a performance note.`
		: null;

	return { rows, note, header, footnote, withheld: false };
}

export function getDefinitions() {
	return DEFINITIONS.map(({ term, definition }) => ({ term, definition }));
}
