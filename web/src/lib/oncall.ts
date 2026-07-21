const HOUR = 3_600_000;
const DAY = 24 * HOUR;

export type PersonTone = 'brand' | 'info' | 'warning' | 'high' | 'success' | 'neutral';

export const PERSON_TONE: Record<string, PersonTone> = {
	'Maya Chen': 'brand',
	'Priya Nair': 'info',
	'Marcus Lee': 'warning',
	'Dev Patel': 'high',
	'Sana Ito': 'success'
};

const TONE_CYCLE: PersonTone[] = ['brand', 'info', 'warning', 'high', 'success'];

export function personTone(name: string): PersonTone {
	if (PERSON_TONE[name]) return PERSON_TONE[name];
	if (!name) return 'neutral';
	let hash = 0;
	for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
	return TONE_CYCLE[hash % TONE_CYCLE.length];
}

// Tailwind cannot see class names built at runtime
export const PERSON_CLASS: Record<PersonTone, string> = {
	brand: 'bg-brand-wash border-brand-edge text-brand-foreground',
	info: 'bg-info-wash border-info-edge text-info-ink',
	warning: 'bg-warning-wash border-warning-edge text-warning-ink',
	high: 'bg-high-wash border-high-edge text-high-ink',
	success: 'bg-success-wash border-success-edge text-success-ink',
	neutral: 'bg-neutral-wash border-neutral-edge text-neutral-ink'
};

export type Rotation = 'daily' | 'weekly' | 'custom';

export const ROTATIONS: { value: Rotation; label: string; description: string }[] = [
	{ value: 'daily', label: 'Daily', description: 'Next person every day at handover' },
	{ value: 'weekly', label: 'Weekly', description: 'Next person every week' },
	{ value: 'custom', label: 'Custom interval', description: 'Every N days' }
];

// Weekday (0 = Sunday), whole hours in the schedule's timezone
export type Restriction = { day: number; start: number; end: number };

export type Layer = {
	id: string;
	participants: string[];
	rotation: Rotation;
	intervalDays: number;
	handoverHour: number;
	startsOn: string;
	restrictions: Restriction[];
};

export type Override = {
	id: string;
	person: string;
	startsAt: string;
	endsAt: string;
	reason: string;
	createdBy: string;
	createdAt: string;
};

export type AuditEntry = { id: string; at: string; by: string; what: string };

export type Schedule = {
	id: string;
	name: string;
	team: string;
	timezone: string;
	// Highest precedence first
	layers: Layer[];
	overrides: Override[];
	feedUrl: string;
	archived: boolean;
	paused: boolean;
};

export function layerName(total: number, index: number): string {
	return `Layer ${total - index}`;
}

export type Coverage = {
	person: string | null;
	via: string | null;
	override: boolean;
};

export type Segment = {
	startsAt: string;
	endsAt: string;
	person: string | null;
	via: string | null;
	override: boolean;
};

export function dayStartHour(schedule: Pick<Schedule, 'layers'>): number {
	return schedule.layers[schedule.layers.length - 1]?.handoverHour ?? 0;
}

export function dayWindow(schedule: Pick<Schedule, 'layers'>, day: Date): { from: Date; to: Date } {
	const from = new Date(day);
	from.setUTCHours(dayStartHour(schedule), 0, 0, 0);
	return { from, to: new Date(from.getTime() + DAY) };
}

// Keep only the segments overlapping [from, to], truncated to that window
export function clipSegments(segments: Segment[], from: Date, to: Date): Segment[] {
	const lo = from.getTime();
	const hi = to.getTime();
	const out: Segment[] = [];
	for (const seg of segments) {
		const start = Date.parse(seg.startsAt);
		const end = Date.parse(seg.endsAt);
		if (end <= lo || start >= hi) continue;
		out.push({
			...seg,
			startsAt: new Date(Math.max(start, lo)).toISOString(),
			endsAt: new Date(Math.min(end, hi)).toISOString()
		});
	}
	return out;
}

export type Handover = { at: string; from: string; to: string };

export type Shift = { startsAt: string; endsAt: string; schedule: string; scheduleId: string };

export type DaySummary = {
	date: string;
	person: string | null;
	override: boolean;
	gap: boolean;
};

// Derive the day's dominant person / override / gap from already-computed segments
export function daySummaryFromSegments(segments: Segment[], date: Date): DaySummary {
	const held = new Map<string, number>();
	for (const run of segments) {
		if (!run.person) continue;
		held.set(run.person, (held.get(run.person) ?? 0) + (Date.parse(run.endsAt) - Date.parse(run.startsAt)));
	}
	const [longest] = [...held.entries()].sort((a, b) => b[1] - a[1]);
	const taken = segments.find((run) => run.override);

	return {
		date: date.toISOString().slice(0, 10),
		person: taken?.person ?? longest?.[0] ?? null,
		override: !!taken,
		gap: segments.some((run) => !run.person)
	};
}

export type Zone = 'utc' | 'local';

function parts(iso: string, zone: Zone): { hour: number; minute: number } {
	const at = new Date(iso);
	return zone === 'utc'
		? { hour: at.getUTCHours(), minute: at.getUTCMinutes() }
		: { hour: at.getHours(), minute: at.getMinutes() };
}

function clock({ hour, minute }: { hour: number; minute: number }): string {
	const hh = String(hour).padStart(2, '0');
	return minute ? `${hh}:${String(minute).padStart(2, '0')}` : hh;
}

export function formatSpanHours(segment: Segment, zone: Zone): string {
	const length = Date.parse(segment.endsAt) - Date.parse(segment.startsAt);
	if (length >= DAY) return 'all day';
	return `${clock(parts(segment.startsAt, zone))}–${clock(parts(segment.endsAt, zone))}`;
}

const WEEKDAY = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

export function weekdayName(iso: string): string {
	return WEEKDAY[new Date(iso).getUTCDay()];
}

export function formatGap(gap: Segment): string {
	const from = new Date(gap.startsAt);
	const to = new Date(gap.endsAt);
	const time = (at: Date) =>
		`${String(at.getUTCHours()).padStart(2, '0')}:${String(at.getUTCMinutes()).padStart(2, '0')}`;

	const sameDay = gap.startsAt.slice(0, 10) === gap.endsAt.slice(0, 10);

	return sameDay
		? `gap ${WEEKDAY[from.getUTCDay()]} ${time(from)}–${time(to)} UTC`
		: `gap ${WEEKDAY[from.getUTCDay()]} ${time(from)} – ${WEEKDAY[to.getUTCDay()]} ${time(to)} UTC`;
}

export function formatShift(shift: { startsAt: string; endsAt: string }, now: number): string {
	const from = new Date(shift.startsAt);
	const to = new Date(shift.endsAt);
	const time = (at: Date) =>
		`${String(at.getUTCHours()).padStart(2, '0')}:${String(at.getUTCMinutes()).padStart(2, '0')}`;

	const today = new Date(now).toISOString().slice(0, 10);
	const tomorrow = new Date(now + DAY).toISOString().slice(0, 10);
	const date = shift.startsAt.slice(0, 10);

	const day =
		date === today
			? 'Today'
			: date === tomorrow
				? 'Tomorrow'
				: `${WEEKDAY[from.getUTCDay()]} ${date}`;

	return date === shift.endsAt.slice(0, 10)
		? `${day} ${time(from)}–${time(to)} UTC`
		: `${day} ${time(from)} → ${WEEKDAY[to.getUTCDay()]} ${time(to)} UTC`;
}

export function formatWhen(iso: string, now: number): string {
	const at = new Date(iso);
	const time = `${String(at.getUTCHours()).padStart(2, '0')}:${String(at.getUTCMinutes()).padStart(2, '0')} UTC`;

	const date = iso.slice(0, 10);
	const today = new Date(now).toISOString().slice(0, 10);
	const tomorrow = new Date(now + DAY).toISOString().slice(0, 10);

	if (date === today) return `today ${time}`;
	if (date === tomorrow) return `tomorrow ${time}`;
	return `${WEEKDAY[at.getUTCDay()]} ${time}`;
}

export function formatDuty(layer: Layer): string {
	if (!layer.restrictions.length) return 'around the clock';

	const byWindow = new Map<string, number[]>();
	for (const window of layer.restrictions) {
		const key = `${window.start}-${window.end}`;
		byWindow.set(key, [...(byWindow.get(key) ?? []), window.day]);
	}

	const hours = (start: number, end: number) =>
		start === 0 && end === 24
			? 'around the clock'
			: `${String(start).padStart(2, '0')}:00–${String(end).padStart(2, '0')}:00`;

	const days = (list: number[]) => {
		const set = new Set(list);
		if (set.size === 7) return 'daily';
		if (set.size === 5 && [1, 2, 3, 4, 5].every((day) => set.has(day))) return 'weekdays';
		if (set.size === 2 && set.has(0) && set.has(6)) return 'weekends';
		return [...set]
			.sort()
			.map((day) => WEEKDAY[day])
			.join(', ');
	};

	return [...byWindow.entries()]
		.map(([key, list]) => {
			const [start, end] = key.split('-').map(Number);
			return `${hours(start, end)}, ${days(list)}`;
		})
		.join(' · ');
}

export function formatRotation(layer: Layer): string {
	if (layer.rotation === 'daily') return 'daily rotation';
	if (layer.rotation === 'weekly') return 'weekly rotation';
	return `every ${layer.intervalDays} days`;
}

export function formatLayerLine(layer: Layer): string {
	const windows = new Set(layer.restrictions.map((window) => `${window.start}-${window.end}`));
	if (windows.size !== 1) return formatRotation(layer);

	const [start, end] = [...windows][0].split('-').map(Number);
	if (start === 0 && end === 24) return formatRotation(layer);

	const hours = `${String(start).padStart(2, '0')}–${String(end).padStart(2, '0')}`;
	const days = new Set(layer.restrictions.map((window) => window.day));

	if (days.size === 7) return `daily ${hours}`;
	if (days.size === 5 && [1, 2, 3, 4, 5].every((day) => days.has(day))) return `weekdays ${hours}`;
	if (days.size === 2 && days.has(0) && days.has(6)) return `weekends ${hours}`;

	return formatRotation(layer);
}

export function initials(name: string): string {
	return name
		.split(' ')
		.map((word) => word[0])
		.join('');
}
