import type { AuditEntry, Layer, Override, Schedule, SwapRequest } from '$lib/oncall';
import { coverageAt, segments } from '$lib/oncall';
import { scenario } from './fixtures';

const HOUR = 3_600_000;
const DAY = 24 * HOUR;

const id = (prefix: string) => prefix + Math.random().toString(36).slice(2, 8);

// UTC Monday of the current week; rotation startsOn anchors to it
export function thisMonday(): string {
	const now = new Date();
	const monday = new Date(
		Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() - ((now.getUTCDay() + 6) % 7))
	);
	return monday.toISOString().slice(0, 10);
}

// Next occurrence of this UTC weekday, strictly after today
function nextWeekday(day: number): Date {
	const at = new Date();
	at.setUTCHours(0, 0, 0, 0);
	at.setUTCDate(at.getUTCDate() + ((day - at.getUTCDay() + 7) % 7 || 7));
	return at;
}

const around = (days: number[]): Layer['restrictions'] =>
	days.map((day) => ({ day, start: 0, end: 24 }));

const between = (days: number[], start: number, end: number): Layer['restrictions'] =>
	days.map((day) => ({ day, start, end }));

const WEEKDAYS = [1, 2, 3, 4, 5];

function seed() {
	const monday = thisMonday();

	const sundayWithAGap = [
		...around([1, 2, 3, 4, 5, 6]),
		{ day: 0, start: 0, end: 18 },
		{ day: 0, start: 22, end: 24 }
	];

	const payments: Schedule = {
		id: 'payments-primary',
		name: 'payments-primary',
		team: 'payments',
		layers: [
			{
				id: 'l-pay-2',
				participants: ['Priya Nair'],
				rotation: 'weekly',
				intervalDays: 7,
				handoverHour: 9,
				startsOn: monday,
				restrictions: between(WEEKDAYS, 9, 18)
			},
			{
				id: 'l-pay-1',
				participants: ['Maya Chen', 'Marcus Lee'],
				rotation: 'weekly',
				intervalDays: 7,
				handoverHour: 9,
				startsOn: monday,
				restrictions: sundayWithAGap
			}
		],
		overrides: [],
		audit: [],
		feedToken: 'f31c9a2e',
		archived: false,
		paused: false
	};

	const saturday = nextWeekday(6);
	payments.overrides.push({
		id: 'ov-1',
		person: 'Sana Ito',
		startsAt: new Date(saturday.getTime() + 9 * HOUR).toISOString(),
		endsAt: new Date(saturday.getTime() + 18 * HOUR).toISOString(),
		reason: "Covering Maya's Saturday",
		createdBy: 'Sana Ito',
		createdAt: new Date(Date.now() - 4 * DAY).toISOString()
	});

	const platform: Schedule = {
		id: 'platform-default',
		name: 'platform-default',
		team: 'platform',
		layers: [
			{
				id: 'l-plat-2',
				participants: ['Priya Nair'],
				rotation: 'weekly',
				intervalDays: 7,
				handoverHour: 9,
				startsOn: monday,
				restrictions: between(WEEKDAYS, 9, 20)
			},
			{
				id: 'l-plat-1',
				participants: ['Dev Patel', 'Sana Ito'],
				rotation: 'weekly',
				intervalDays: 7,
				handoverHour: 9,
				startsOn: monday,
				restrictions: []
			}
		],
		overrides: [],
		audit: [],
		feedToken: 'a7d40b16',
		archived: false,
		paused: false
	};

	const thursday = nextWeekday(4);
	platform.overrides.push({
		id: 'ov-2',
		person: 'Maya Chen',
		startsAt: new Date(thursday.getTime() + 18 * HOUR).toISOString(),
		endsAt: new Date(thursday.getTime() + 33 * HOUR).toISOString(),
		reason: 'Platform cover',
		createdBy: 'Priya Nair',
		createdAt: new Date(Date.now() - 2 * DAY).toISOString()
	});

	const frontend: Schedule = {
		id: 'frontend-daytime',
		name: 'frontend-daytime',
		team: 'frontend',
		layers: [
			{
				id: 'l-front-1',
				participants: ['Dev Patel', 'Sana Ito'],
				rotation: 'daily',
				intervalDays: 1,
				handoverHour: 9,
				startsOn: monday,
				restrictions: []
			}
		],
		overrides: [],
		audit: [],
		feedToken: 'c904e57b',
		archived: false,
		paused: false
	};

	payments.audit = [
		{
			id: 'a-1',
			at: new Date(Date.now() - 4 * DAY).toISOString(),
			by: 'Sana Ito',
			what: `Added override: Sana Ito takes Sat ${saturday.toISOString().slice(0, 10)} 09:00–18:00 UTC — covering Maya's Saturday`
		},
		{
			id: 'a-2',
			at: new Date(Date.now() - 32 * DAY).toISOString(),
			by: 'Priya Nair',
			what: 'Added a weekday daytime layer above the rotation'
		},
		{
			id: 'a-3',
			at: new Date(Date.now() - 72 * DAY).toISOString(),
			by: 'Maya Chen',
			what: 'Created the schedule with one weekly rotation'
		}
	];

	const requests: SwapRequest[] = [
		{
			id: 'rq-1',
			text: `Swap ${nextWeekday(0).toISOString().slice(0, 10)} day shift with Dev Patel`,
			message: 'Family thing on Sunday — can you take the day part?',
			status: 'pending'
		},
		{
			id: 'rq-2',
			text: 'Handed last Saturday day shift to Sana Ito',
			message: '',
			status: 'approved'
		}
	];

	return { schedules: [payments, platform, frontend], requests };
}

const store = seed();

if (scenario() === 'empty') {
	store.schedules.length = 0;
	store.requests.length = 0;
}

function record(schedule: Schedule, what: string, by = 'Maya Chen') {
	const entry: AuditEntry = { id: id('a-'), at: new Date().toISOString(), by, what };
	schedule.audit.unshift(entry);
}

export function listSchedules(): Schedule[] {
	return store.schedules.filter((schedule) => !schedule.archived);
}

export function getSchedule(scheduleId: string): Schedule | undefined {
	return store.schedules.find((schedule) => schedule.id === scheduleId);
}

// Schedule names are URL segments; these collide with routes
const RESERVED = ['new', 'mine'];

export function nameTaken(name: string, exceptId?: string): boolean {
	if (RESERVED.includes(name)) return true;
	return store.schedules.some((schedule) => schedule.id === name && schedule.id !== exceptId);
}

function freeId(base: string): string {
	if (!nameTaken(base)) return base;
	for (let suffix = 2; ; suffix++) {
		const candidate = `${base}-${suffix}`;
		if (!nameTaken(candidate)) return candidate;
	}
}

export type ScheduleInput = {
	name: string;
	team: string;
	layers: Layer[];
};

export function createSchedule(input: ScheduleInput, by = 'Maya Chen'): Schedule {
	const schedule: Schedule = {
		id: input.name,
		name: input.name,
		team: input.team,
		layers: input.layers,
		overrides: [],
		audit: [],
		feedToken: Math.random().toString(16).slice(2, 10),
		archived: false,
		paused: false
	};

	record(schedule, `Created the schedule with ${input.layers.length} layer(s)`, by);

	store.schedules.push(schedule);
	return schedule;
}

export function updateSchedule(scheduleId: string, input: ScheduleInput, by = 'Maya Chen') {
	const schedule = getSchedule(scheduleId);
	if (!schedule) return;

	const renamed = input.name !== schedule.name;

	schedule.id = input.name;
	schedule.name = input.name;
	schedule.team = input.team;
	schedule.layers = input.layers;

	record(
		schedule,
		renamed
			? `Renamed to ${input.name}. The old calendar feed URL stops working.`
			: `Edited the schedule — ${input.layers.length} layer(s)`,
		by
	);
}

export function duplicateSchedule(scheduleId: string, by = 'Maya Chen'): Schedule | undefined {
	const schedule = getSchedule(scheduleId);
	if (!schedule) return;

	const id = freeId(`${schedule.id}-copy`);

	const copy: Schedule = {
		...structuredClone(schedule),
		id,
		name: id,
		overrides: [],
		audit: [],
		feedToken: Math.random().toString(16).slice(2, 10),
		paused: true
	};

	record(copy, `Duplicated from ${schedule.name}. It starts paused.`, by);
	store.schedules.push(copy);
	return copy;
}

export function resumeSchedule(scheduleId: string, by = 'Maya Chen') {
	const schedule = getSchedule(scheduleId);
	if (!schedule) return;

	schedule.paused = false;
	record(schedule, 'Resumed the schedule. It pages again from the next handover.', by);
}

export function archiveSchedule(scheduleId: string, by = 'Maya Chen') {
	const schedule = getSchedule(scheduleId);
	if (!schedule) return;

	schedule.archived = true;
	record(schedule, 'Archived the schedule', by);
}

export type OverrideInput = {
	person: string;
	startsAt: string;
	endsAt: string;
	reason: string;
};

export function addOverride(
	scheduleId: string,
	input: OverrideInput,
	by = 'Maya Chen'
): { error: string } | { override: Override } {
	const schedule = getSchedule(scheduleId);
	if (!schedule) return { error: 'That schedule no longer exists.' };

	const from = new Date(input.startsAt);
	const to = new Date(input.endsAt);
	if (!(to > from)) return { error: 'The override has to end after it starts.' };

	const covered = segments(schedule, from, to);
	if (covered.every((run) => run.person === input.person)) {
		return {
			error: `${input.person} already holds this shift. Pick someone else, or change the window.`
		};
	}

	const override: Override = {
		id: id('ov-'),
		person: input.person,
		startsAt: from.toISOString(),
		endsAt: to.toISOString(),
		reason: input.reason,
		createdBy: by,
		createdAt: new Date().toISOString()
	};

	schedule.overrides.push(override);

	const window = `${from.toISOString().slice(0, 16).replace('T', ' ')}–${to.toISOString().slice(11, 16)} UTC`;
	record(schedule, `Added override: ${input.person} takes ${window} — ${override.reason}`, by);

	return { override };
}

export function onCallNow(schedule: Schedule, now: Date) {
	if (schedule.paused) return { person: null, until: null };

	const cover = coverageAt(schedule, now);
	if (!cover.person) return { person: null, until: null };

	const [run] = segments(schedule, now, new Date(now.getTime() + 14 * DAY));
	return { person: cover.person, until: run?.endsAt ?? null };
}

export function listSwapRequests(): SwapRequest[] {
	return store.requests;
}

export function requestSwap(text: string, message: string) {
	store.requests.unshift({ id: id('rq-'), text, message, status: 'pending' });
}
