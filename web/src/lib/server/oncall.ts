import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import {
	clipSegments,
	daySummaryFromSegments,
	dayWindow,
	formatDuty,
	formatGap,
	layerName,
	type Handover,
	type Layer,
	type Override,
	type Schedule,
	type Segment
} from '$lib/oncall';
import { apiClient } from './api';

type Schemas = components['schemas'];

const DAY = 86_400_000;
const WEEKDAY = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

// UTC Monday of the current week; the create form anchors previews to it
export function thisMonday(): string {
	const now = new Date();
	const monday = new Date(
		Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() - ((now.getUTCDay() + 6) % 7))
	);
	return monday.toISOString().slice(0, 10);
}

type MemberIndex = { byId: Map<string, string>; byName: Map<string, string> };

async function memberIndex(cookies: Cookies, workspace: string): Promise<MemberIndex> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/members', {
		params: { path: { workspaceId: workspace } }
	});
	const members = data?.items ?? [];
	return {
		byId: new Map(members.map((m) => [m.userId, m.name])),
		byName: new Map(members.map((m) => [m.name, m.userId]))
	};
}

async function teamSlugs(cookies: Cookies, workspace: string): Promise<string[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/teams', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map((t) => t.slug);
}

function name(byId: Map<string, string>, id?: string | null): string {
	if (!id) return '';
	return byId.get(id) ?? id;
}

function toSegment(dto: Schemas['Segment'], byId: Map<string, string>): Segment {
	return {
		startsAt: dto.startsAt,
		endsAt: dto.endsAt,
		person: dto.userId ? name(byId, dto.userId) : null,
		via: dto.via ?? null,
		override: dto.override
	};
}

function toHandover(dto: Schemas['Handover'], byId: Map<string, string>): Handover {
	return { at: dto.at, from: name(byId, dto.fromUserId), to: name(byId, dto.toUserId) };
}

function toOverride(dto: Schemas['ScheduleOverride'], byId: Map<string, string>): Override {
	return {
		id: dto.id,
		person: name(byId, dto.userId),
		startsAt: dto.startsAt,
		endsAt: dto.endsAt,
		reason: dto.reason,
		createdBy: name(byId, dto.createdByUserId),
		createdAt: dto.createdAt
	};
}

function toLayer(dto: Schemas['ScheduleLayer'], byId: Map<string, string>): Layer {
	return {
		id: dto.id,
		participants: dto.participants.map((id) => name(byId, id)),
		rotation: dto.rotation,
		intervalDays: dto.intervalDays,
		handoverHour: dto.handoverHour,
		startsOn: dto.startsOn,
		restrictions: dto.restrictions.map((r) => ({
			day: r.weekday,
			start: r.startMinute / 60,
			end: r.endMinute / 60
		}))
	};
}

function toSchedule(dto: Schemas['Schedule'], byId: Map<string, string>): Schedule {
	return {
		id: dto.slug,
		name: dto.slug,
		team: dto.team,
		timezone: dto.timezone,
		layers: dto.layers.map((l) => toLayer(l, byId)),
		overrides: dto.overrides.map((o) => toOverride(o, byId)),
		feedUrl: dto.feedUrl,
		archived: dto.archived,
		paused: dto.paused
	};
}

type ApiLayer = Schemas['ScheduleLayerInput'];

function toApiLayers(layers: Layer[], byName: Map<string, string>): ApiLayer[] {
	return layers.map((layer) => ({
		participants: layer.participants
			.map((n) => byName.get(n))
			.filter((id): id is string => Boolean(id)),
		rotation: layer.rotation,
		intervalDays: layer.intervalDays,
		handoverHour: layer.handoverHour,
		startsOn: layer.startsOn,
		restrictions: layer.restrictions.map((r) => ({
			weekday: r.day,
			startMinute: Math.round(r.start * 60),
			endMinute: Math.round(r.end * 60)
		}))
	}));
}

const utcDay = (at: Date) => new Date(Date.UTC(at.getUTCFullYear(), at.getUTCMonth(), at.getUTCDate()));

export type ScheduleInput = { name: string; team: string; timezone: string; layers: Layer[] };
export type OverrideInput = { person: string; startsAt: string; endsAt: string; reason: string };

// ---- List ----------------------------------------------------------------

export async function scheduleList(cookies: Cookies, workspace: string, includeArchived = false) {
	const client = apiClient(cookies);
	const [listRes, index] = await Promise.all([
		client.GET('/workspaces/{workspaceId}/schedules', {
			params: {
				path: { workspaceId: workspace },
				query: { includeArchived: includeArchived || undefined }
			}
		}),
		memberIndex(cookies, workspace)
	]);
	const items = listRes.data?.items ?? [];
	const now = Date.now();
	const from = new Date(now).toISOString();
	const to = new Date(now + 14 * DAY).toISOString();

	const schedules = await Promise.all(
		items.map(async (s) => {
			const base = {
				id: s.slug,
				name: s.slug,
				team: s.team,
				paused: s.paused,
				archived: s.archived
			};
			if (s.archived) {
				return { ...base, gap: null, handover: null, person: null, until: null };
			}
			const cal = await client.GET(
				'/workspaces/{workspaceId}/schedules/{scheduleSlug}/calendar',
				{ params: { path: { workspaceId: workspace, scheduleSlug: s.slug }, query: { from, to } } }
			);
			const gaps = (cal.data?.gaps ?? []).map((seg) => toSegment(seg, index.byId));
			const handovers = (cal.data?.handovers ?? []).map((h) => toHandover(h, index.byId));
			return {
				...base,
				gap: gaps[0] ? formatGap(gaps[0]) : null,
				handover: handovers[0] ?? null,
				person: s.onCallUserId ? name(index.byId, s.onCallUserId) : null,
				until: s.onCallUntil ?? null
			};
		})
	);
	return { now, schedules };
}

// ---- Detail --------------------------------------------------------------

type DetailOptions = { view: 'week' | 'month'; zone: 'utc' | 'local'; date?: string; time?: string };

export async function scheduleDetail(
	cookies: Cookies,
	workspace: string,
	slug: string,
	opts: DetailOptions
) {
	const client = apiClient(cookies);
	const path = { workspaceId: workspace, scheduleSlug: slug };
	const now = new Date();

	const weekStart = utcDay(new Date(now.getTime() - DAY));
	const weekRange = { from: weekStart.toISOString(), to: new Date(weekStart.getTime() + 8 * DAY).toISOString() };
	const monthStart = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), 1));
	const monthLength = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 0)).getUTCDate();
	const monthRange = {
		from: monthStart.toISOString(),
		to: new Date(monthStart.getTime() + (monthLength + 1) * DAY).toISOString()
	};
	const upcomingRange = { from: now.toISOString(), to: new Date(now.getTime() + 14 * DAY).toISOString() };

	const date = opts.date ?? now.toISOString().slice(0, 10);
	const time = opts.time ?? `${String(now.getUTCHours()).padStart(2, '0')}:00`;
	const at = new Date(`${date}T${time}:00Z`);

	const [schedRes, weekCal, monthCal, upcomingCal, onCallRes, index] = await Promise.all([
		client.GET('/workspaces/{workspaceId}/schedules/{scheduleSlug}', { params: { path } }),
		client.GET('/workspaces/{workspaceId}/schedules/{scheduleSlug}/calendar', { params: { path, query: weekRange } }),
		client.GET('/workspaces/{workspaceId}/schedules/{scheduleSlug}/calendar', { params: { path, query: monthRange } }),
		client.GET('/workspaces/{workspaceId}/schedules/{scheduleSlug}/calendar', { params: { path, query: upcomingRange } }),
		client.GET('/workspaces/{workspaceId}/schedules/{scheduleSlug}/on-call', {
			params: { path, query: { at: Number.isNaN(at.getTime()) ? now.toISOString() : at.toISOString() } }
		}),
		memberIndex(cookies, workspace)
	]);
	if (!schedRes.data) return null;

	const schedule = toSchedule(schedRes.data, index.byId);
	const weekEffective = (weekCal.data?.effective ?? []).map((s) => toSegment(s, index.byId));
	const weekLayers = (weekCal.data?.layers ?? []).map((l) => ({
		index: l.index,
		segments: l.segments.map((s) => toSegment(s, index.byId))
	}));
	const monthEffective = (monthCal.data?.effective ?? []).map((s) => toSegment(s, index.byId));
	const upcomingEffective = (upcomingCal.data?.effective ?? []).map((s) => toSegment(s, index.byId));
	const handovers = (upcomingCal.data?.handovers ?? []).map((h) => toHandover(h, index.byId)).slice(0, 8);
	const gaps = (upcomingCal.data?.gaps ?? []).map((s) => toSegment(s, index.byId));

	const week = Array.from({ length: 7 }, (_, i) => {
		const day = new Date(weekStart.getTime() + i * DAY);
		const { from, to } = dayWindow(schedule, day);
		return {
			label: WEEKDAY[day.getUTCDay()],
			num: day.getUTCDate(),
			today: day.getTime() === utcDay(now).getTime(),
			from,
			to
		};
	});

	const monthDays = Array.from({ length: monthLength }, (_, i) => {
		const day = new Date(monthStart.getTime() + i * DAY);
		const { from, to } = dayWindow(schedule, day);
		return daySummaryFromSegments(clipSegments(monthEffective, from, to), from);
	});

	const target =
		upcomingEffective.find((run) => Date.parse(run.startsAt) > now.getTime()) ?? upcomingEffective[0];

	const onCall = onCallRes.data;
	const coverage = {
		person: onCall?.userId ? name(index.byId, onCall.userId) : null,
		via: onCall?.via ?? null,
		override: onCall?.override ?? false
	};

	return {
		now: now.getTime(),
		view: opts.view,
		zone: opts.zone,
		id: schedule.id,
		name: schedule.name,
		team: schedule.team,
		timezone: schedule.timezone,
		paused: schedule.paused,
		archived: schedule.archived,
		gap: gaps[0] ? formatGap(gaps[0]) : null,
		weekLabel: `${week[0].label} ${week[0].num} – ${week[6].label} ${week[6].num}`,
		days: week.map(({ label, num, today }) => ({ label, num, today })),
		effective: week.map((day) => clipSegments(weekEffective, day.from, day.to)),
		reasons: Object.fromEntries(schedule.overrides.map((o) => [o.startsAt, o.reason])),
		layers: schedule.layers.map((layer, i) => ({
			layer,
			name: layerName(schedule.layers.length, i),
			duty: formatDuty(layer),
			days: week.map((day) =>
				clipSegments(weekLayers.find((l) => l.index === i)?.segments ?? [], day.from, day.to)
			)
		})),
		month: {
			blanks: (monthStart.getUTCDay() + 6) % 7,
			label: monthStart.toLocaleDateString('en-GB', { month: 'long', year: 'numeric', timeZone: 'UTC' }),
			days: monthDays
		},
		handovers,
		resolver: { date, time, coverage },
		target: target
			? { startsAt: target.startsAt, endsAt: target.endsAt, person: target.person }
			: {
					startsAt: now.toISOString(),
					endsAt: new Date(now.getTime() + DAY).toISOString(),
					person: null
				},
		people: [...index.byId.values()].sort((a, b) => a.localeCompare(b)),
		feedUrl: schedule.feedUrl,
		audit: await scheduleAudit(cookies, workspace, slug)
	};
}

const AUDIT_TEXT: Record<string, string> = {
	'schedule.created': 'Created the schedule',
	'schedule.updated': 'Edited the schedule',
	'schedule.duplicated': 'Duplicated the schedule',
	'schedule.archived': 'Archived the schedule',
	'schedule.unarchived': 'Restored the schedule',
	'schedule.deleted': 'Deleted the schedule',
	'schedule.paused': 'Paused the schedule',
	'schedule.resumed': 'Resumed the schedule',
	'schedule.override.added': 'Added an override',
	'schedule.reassigned': 'Reassigned a participant'
};

async function scheduleAudit(cookies: Cookies, workspace: string, slug: string) {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/audit', {
		params: { path: { workspaceId: workspace }, query: { action: 'schedule', limit: 50 } }
	});
	return (data?.items ?? [])
		.filter((entry) => entry.target === slug)
		.map((entry) => ({
			id: entry.id,
			at: entry.at,
			by: entry.actor || 'system',
			what: AUDIT_TEXT[entry.action] ?? entry.action
		}));
}

// ---- Form data (new / edit) ----------------------------------------------

export async function formOptions(cookies: Cookies, workspace: string) {
	const [index, teams] = await Promise.all([memberIndex(cookies, workspace), teamSlugs(cookies, workspace)]);
	return { people: [...index.byId.values()].sort((a, b) => a.localeCompare(b)), teams };
}

export async function editSchedule(cookies: Cookies, workspace: string, slug: string) {
	const [{ data }, index] = await Promise.all([
		apiClient(cookies).GET('/workspaces/{workspaceId}/schedules/{scheduleSlug}', {
			params: { path: { workspaceId: workspace, scheduleSlug: slug } }
		}),
		memberIndex(cookies, workspace)
	]);
	if (!data) return null;
	const schedule = toSchedule(data, index.byId);
	return { name: schedule.name, team: schedule.team, timezone: schedule.timezone, layers: schedule.layers };
}

// ---- Preview -------------------------------------------------------------

export async function previewRows(cookies: Cookies, workspace: string, input: ScheduleInput, from: string) {
	const { byName, byId } = await memberIndex(cookies, workspace);
	const fromDate = new Date(`${from}T00:00:00Z`);
	const to = new Date(fromDate.getTime() + 7 * DAY);

	const { data } = await apiClient(cookies).POST('/workspaces/{workspaceId}/schedules/preview', {
		params: { path: { workspaceId: workspace } },
		body: {
			timezone: input.timezone,
			layers: toApiLayers(input.layers, byName),
			from: fromDate.toISOString(),
			to: to.toISOString()
		}
	});

	const effectiveSegments = (data?.effective ?? []).map((s) => toSegment(s, byId));
	const layerSegments = (data?.layers ?? []).map((l) => ({
		index: l.index,
		segments: l.segments.map((s) => toSegment(s, byId))
	}));

	const dayStart = input.layers[input.layers.length - 1]?.handoverHour ?? 0;
	const days = Array.from({ length: 7 }, (_, i) => new Date(fromDate.getTime() + i * DAY));
	const windowOf = (day: Date) => {
		const f = new Date(day);
		f.setUTCHours(dayStart, 0, 0, 0);
		return { from: f, to: new Date(f.getTime() + DAY) };
	};

	return {
		days: days.map((d) => ({ label: WEEKDAY[d.getUTCDay()], num: d.getUTCDate(), iso: d.toISOString() })),
		effective: days.map((d) => {
			const w = windowOf(d);
			return daySummaryFromSegments(clipSegments(effectiveSegments, w.from, w.to), w.from);
		}),
		rows: input.layers.map((layer, i) => ({
			label: `L${input.layers.length - i}`,
			title: layerName(input.layers.length, i),
			days: days.map((d) => {
				const w = windowOf(d);
				const segs = layerSegments.find((l) => l.index === i)?.segments ?? [];
				return daySummaryFromSegments(clipSegments(segs, w.from, w.to), w.from);
			})
		}))
	};
}

// ---- Mine ----------------------------------------------------------------

export async function myShifts(cookies: Cookies, workspace: string, from: string, to: string) {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/on-call', {
		params: { path: { workspaceId: workspace }, query: { from, to } }
	});
	return (data?.items ?? []).map((shift) => ({
		startsAt: shift.startsAt,
		endsAt: shift.endsAt,
		scheduleSlug: shift.scheduleSlug
	}));
}

// ---- Mutations -----------------------------------------------------------

type SaveResult = { slug?: string; nameError?: string; error?: string };

export async function createSchedule(
	cookies: Cookies,
	workspace: string,
	input: ScheduleInput
): Promise<SaveResult> {
	const { byName } = await memberIndex(cookies, workspace);
	const { data, error, response } = await apiClient(cookies).POST('/workspaces/{workspaceId}/schedules', {
		params: { path: { workspaceId: workspace } },
		body: { name: input.name, team: input.team, timezone: input.timezone, layers: toApiLayers(input.layers, byName) }
	});
	if (data) return { slug: data.slug };
	if (response?.status === 409) return { nameError: error?.detail ?? 'A schedule already goes by that name.' };
	return { error: error?.detail ?? 'Could not create the schedule.' };
}

export async function updateSchedule(
	cookies: Cookies,
	workspace: string,
	slug: string,
	input: ScheduleInput
): Promise<SaveResult> {
	const { byName } = await memberIndex(cookies, workspace);
	const { data, error, response } = await apiClient(cookies).PUT(
		'/workspaces/{workspaceId}/schedules/{scheduleSlug}',
		{
			params: { path: { workspaceId: workspace, scheduleSlug: slug } },
			body: { name: input.name, team: input.team, timezone: input.timezone, layers: toApiLayers(input.layers, byName) }
		}
	);
	if (data) return { slug: data.slug };
	if (response?.status === 409) return { nameError: error?.detail ?? 'A schedule already goes by that name.' };
	return { error: error?.detail ?? 'Could not save the schedule.' };
}

export async function duplicateSchedule(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<{ slug?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/schedules/{scheduleSlug}/duplicate',
		{ params: { path: { workspaceId: workspace, scheduleSlug: slug } } }
	);
	if (data) return { slug: data.slug };
	return { error: error?.detail ?? 'Could not duplicate the schedule.' };
}

async function scheduleAction(
	cookies: Cookies,
	workspace: string,
	slug: string,
	action: 'archive' | 'unarchive' | 'pause' | 'resume'
): Promise<boolean> {
	const paths = {
		archive: '/workspaces/{workspaceId}/schedules/{scheduleSlug}/archive',
		unarchive: '/workspaces/{workspaceId}/schedules/{scheduleSlug}/unarchive',
		pause: '/workspaces/{workspaceId}/schedules/{scheduleSlug}/pause',
		resume: '/workspaces/{workspaceId}/schedules/{scheduleSlug}/resume'
	} as const;
	const { error } = await apiClient(cookies).POST(paths[action], {
		params: { path: { workspaceId: workspace, scheduleSlug: slug } }
	});
	return !error;
}

export const archiveSchedule = (c: Cookies, w: string, s: string) => scheduleAction(c, w, s, 'archive');
export const unarchiveSchedule = (c: Cookies, w: string, s: string) => scheduleAction(c, w, s, 'unarchive');
export const pauseSchedule = (c: Cookies, w: string, s: string) => scheduleAction(c, w, s, 'pause');
export const resumeSchedule = (c: Cookies, w: string, s: string) => scheduleAction(c, w, s, 'resume');

export async function deleteSchedule(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).DELETE('/workspaces/{workspaceId}/schedules/{scheduleSlug}', {
		params: { path: { workspaceId: workspace, scheduleSlug: slug } }
	});
	return error ? { error: error.detail ?? 'Could not delete the schedule.' } : {};
}

export async function addOverride(
	cookies: Cookies,
	workspace: string,
	slug: string,
	input: OverrideInput
): Promise<{ error?: string }> {
	const { byName } = await memberIndex(cookies, workspace);
	const userId = byName.get(input.person);
	if (!userId) return { error: 'Pick a person from the workspace.' };
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/schedules/{scheduleSlug}/overrides',
		{
			params: { path: { workspaceId: workspace, scheduleSlug: slug } },
			body: { userId, startsAt: input.startsAt, endsAt: input.endsAt, reason: input.reason }
		}
	);
	return error ? { error: error.detail ?? 'Could not add the override.' } : {};
}
