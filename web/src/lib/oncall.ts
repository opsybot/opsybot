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

export function personTone(name: string): PersonTone {
	return PERSON_TONE[name] ?? 'neutral';
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

export const PEOPLE = ['Maya Chen', 'Priya Nair', 'Marcus Lee', 'Dev Patel', 'Sana Ito'];

export const UNREACHABLE = ['Jordan Okafor'];

export const ASSIGNABLE = [...PEOPLE, ...UNREACHABLE];

export const TEAMS = ['payments', 'platform', 'frontend'];

export type Rotation = 'daily' | 'weekly' | 'custom';

export const ROTATIONS: { value: Rotation; label: string; description: string }[] = [
	{ value: 'daily', label: 'Daily', description: 'Next person every day at handover' },
	{ value: 'weekly', label: 'Weekly', description: 'Next person every week' },
	{ value: 'custom', label: 'Custom interval', description: 'Every N days' }
];

// UTC weekday (0 = Sunday), whole UTC hours
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
	// Highest precedence first
	layers: Layer[];
	overrides: Override[];
	audit: AuditEntry[];
	feedToken: string;
	archived: boolean;
	paused: boolean;
};

export type SwapRequest = {
	id: string;
	text: string;
	message: string;
	status: 'pending' | 'approved';
};

export function layerName(total: number, index: number): string {
	return `Layer ${total - index}`;
}

export function rotationPerson(layer: Layer, at: Date): string | null {
	if (!layer.participants.length) return null;

	const period =
		layer.rotation === 'daily' ? 1 : layer.rotation === 'weekly' ? 7 : Math.max(1, layer.intervalDays);

	const anchor = Date.parse(
		`${layer.startsOn}T${String(layer.handoverHour).padStart(2, '0')}:00:00Z`
	);
	const periods = Math.floor((at.getTime() - anchor) / (period * DAY));
	const count = layer.participants.length;

	return layer.participants[((periods % count) + count) % count];
}

export function layerOnDuty(layer: Layer, at: Date): boolean {
	if (!layer.restrictions.length) return true;

	const day = at.getUTCDay();
	const hour = at.getUTCHours() + at.getUTCMinutes() / 60;

	return layer.restrictions.some(
		(window) => window.day === day && hour >= window.start && hour < window.end
	);
}

export type Coverage = {
	person: string | null;
	via: string | null;
	override: boolean;
};

const NOBODY: Coverage = { person: null, via: null, override: false };

export function coverageAt(schedule: Schedule, at: Date): Coverage {
	const instant = at.getTime();

	// The last override written wins
	const override = schedule.overrides.findLast(
		(entry) => instant >= Date.parse(entry.startsAt) && instant < Date.parse(entry.endsAt)
	);
	if (override) return { person: override.person, via: 'override', override: true };

	for (let index = 0; index < schedule.layers.length; index++) {
		const cover = layerCoverageAt(schedule, index, at);
		if (cover.person) return cover;
	}

	return NOBODY;
}

function layerCoverageAt(schedule: Schedule, index: number, at: Date): Coverage {
	const layer = schedule.layers[index];
	if (!layer || !layerOnDuty(layer, at)) return NOBODY;

	const person = rotationPerson(layer, at);
	if (!person) return NOBODY;

	return {
		person,
		via: layerName(schedule.layers.length, index).toLowerCase(),
		override: false
	};
}

export type Segment = {
	startsAt: string;
	endsAt: string;
	person: string | null;
	via: string | null;
	override: boolean;
};

function boundaries(schedule: Schedule, from: number, to: number): number[] {
	const marks = new Set<number>([from, to]);

	for (let mark = Math.ceil(from / HOUR) * HOUR; mark < to; mark += HOUR) marks.add(mark);

	for (const override of schedule.overrides) {
		for (const mark of [Date.parse(override.startsAt), Date.parse(override.endsAt)]) {
			if (mark > from && mark < to) marks.add(mark);
		}
	}

	return [...marks].sort((a, b) => a - b);
}

export function segments(
	schedule: Schedule,
	from: Date,
	to: Date,
	layerIndex?: number
): Segment[] {
	const marks = boundaries(schedule, from.getTime(), to.getTime());
	const out: Segment[] = [];

	for (let i = 0; i < marks.length - 1; i++) {
		const [start, end] = [marks[i], marks[i + 1]];
		const at = new Date((start + end) / 2);
		const cover =
			layerIndex === undefined
				? coverageAt(schedule, at)
				: layerCoverageAt(schedule, layerIndex, at);

		const last = out[out.length - 1];
		if (last && last.person === cover.person && last.override === cover.override) {
			last.endsAt = new Date(end).toISOString();
			continue;
		}

		out.push({
			startsAt: new Date(start).toISOString(),
			endsAt: new Date(end).toISOString(),
			...cover
		});
	}

	return out;
}

export function dayStartHour(schedule: Schedule): number {
	return schedule.layers[schedule.layers.length - 1]?.handoverHour ?? 0;
}

export function dayWindow(schedule: Schedule, day: Date): { from: Date; to: Date } {
	const from = new Date(day);
	from.setUTCHours(dayStartHour(schedule), 0, 0, 0);
	return { from, to: new Date(from.getTime() + DAY) };
}

export function gaps(schedule: Schedule, from: Date, days: number): Segment[] {
	if (schedule.paused) return [];

	const to = new Date(from.getTime() + days * DAY);
	return segments(schedule, from, to).filter((segment) => !segment.person);
}

export type Handover = { at: string; from: string; to: string };

export function handovers(schedule: Schedule, from: Date, days: number, limit = 3): Handover[] {
	if (schedule.paused) return [];

	const to = new Date(from.getTime() + days * DAY);
	const runs = segments(schedule, from, to);
	const out: Handover[] = [];

	for (let i = 1; i < runs.length && out.length < limit; i++) {
		const [before, after] = [runs[i - 1], runs[i]];
		if (!before.person || !after.person || before.person === after.person) continue;
		out.push({ at: after.startsAt, from: before.person, to: after.person });
	}

	return out;
}

export type Shift = { startsAt: string; endsAt: string; schedule: string; scheduleId: string };

export function shiftsFor(
	schedules: Schedule[],
	person: string,
	from: Date,
	days: number
): Shift[] {
	const to = new Date(from.getTime() + days * DAY);

	// Look past the window end so a straddling shift keeps its real end time
	const horizon = new Date(to.getTime() + 7 * DAY);

	return schedules
		.filter((schedule) => !schedule.paused)
		.flatMap((schedule) =>
			segments(schedule, from, horizon)
				.filter(
					(segment) =>
						segment.person === person && Date.parse(segment.startsAt) < to.getTime()
				)
				.map((segment) => ({
					startsAt: segment.startsAt,
					endsAt: segment.endsAt,
					schedule: schedule.name,
					scheduleId: schedule.id
				}))
		)
		.sort((a, b) => Date.parse(a.startsAt) - Date.parse(b.startsAt));
}

export type DaySummary = {
	date: string;
	person: string | null;
	override: boolean;
	gap: boolean;
};

export function daySummary(schedule: Schedule, day: Date): DaySummary {
	const { from, to } = dayWindow(schedule, day);
	const runs = segments(schedule, from, to);

	const held = new Map<string, number>();
	for (const run of runs) {
		if (!run.person) continue;
		const length = Date.parse(run.endsAt) - Date.parse(run.startsAt);
		held.set(run.person, (held.get(run.person) ?? 0) + length);
	}

	const [longest] = [...held.entries()].sort((a, b) => b[1] - a[1]);

	const taken = runs.find((run) => run.override);

	return {
		date: from.toISOString().slice(0, 10),
		person: taken?.person ?? longest?.[0] ?? null,
		override: !!taken,
		gap: runs.some((run) => !run.person)
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
			: `${String(start).padStart(2, '0')}:00–${String(end).padStart(2, '0')}:00 UTC`;

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
