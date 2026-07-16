import { error, fail, redirect } from '@sveltejs/kit';
import {
	coverageAt,
	dayWindow,
	daySummary,
	formatDuty,
	formatGap,
	gaps,
	handovers,
	layerName,
	segments,
	type Zone
} from '$lib/oncall';
import {
	addOverride,
	archiveSchedule,
	duplicateSchedule,
	getSchedule,
	resumeSchedule
} from '$lib/server/oncall';
import type { Actions, PageServerLoad } from './$types';

const DAY = 86_400_000;

// Self-hosted instances point this at their own host
const FEED_ORIGIN = 'https://opsy.bot';

const WEEKDAY = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

const utcDay = (at: Date) => new Date(Date.UTC(at.getUTCFullYear(), at.getUTCMonth(), at.getUTCDate()));

export const load: PageServerLoad = ({ params, url }) => {
	const schedule = getSchedule(params.id);
	if (!schedule) error(404, `No schedule called ${params.id}.`);

	const now = new Date();
	const view = url.searchParams.get('view') === 'month' ? 'month' : 'week';
	const zone: Zone = url.searchParams.get('tz') === 'local' ? 'local' : 'utc';

	// Week starts a day back so the shift that just ended stays visible
	const first = utcDay(new Date(now.getTime() - DAY));
	const week = Array.from({ length: 7 }, (_, index) => {
		const date = new Date(first.getTime() + index * DAY);
		const { from, to } = dayWindow(schedule, date);
		return {
			date: date.toISOString().slice(0, 10),
			label: WEEKDAY[date.getUTCDay()],
			num: date.getUTCDate(),
			today: date.getTime() === utcDay(now).getTime(),
			from,
			to
		};
	});

	const month = (() => {
		const start = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), 1));
		const length = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 0)).getUTCDate();

		return {
			blanks: (start.getUTCDay() + 6) % 7,
			label: start.toLocaleDateString('en-GB', { month: 'long', year: 'numeric', timeZone: 'UTC' }),
			days: Array.from({ length }, (_, index) =>
				daySummary(schedule, new Date(start.getTime() + index * DAY))
			)
		};
	})();

	const date = url.searchParams.get('date') ?? now.toISOString().slice(0, 10);
	const time = url.searchParams.get('time') ?? `${String(now.getUTCHours()).padStart(2, '0')}:00`;
	const at = new Date(`${date}T${time}:00Z`);

	const upcoming = segments(schedule, now, new Date(now.getTime() + 14 * DAY));
	const target = upcoming.find((run) => Date.parse(run.startsAt) > now.getTime()) ?? upcoming[0];

	const [gap] = gaps(schedule, now, 7);

	return {
		now: now.getTime(),
		view,
		zone,
		id: schedule.id,
		name: schedule.name,
		team: schedule.team,
		paused: schedule.paused,
		archived: schedule.archived,
		gap: gap ? formatGap(gap) : null,
		weekLabel: `${week[0].label} ${week[0].num} – ${week[6].label} ${week[6].num}`,
		days: week.map(({ label, num, today }) => ({ label, num, today })),
		effective: week.map((day) => segments(schedule, day.from, day.to)),
		reasons: Object.fromEntries(
			schedule.overrides.map((override) => [override.startsAt, override.reason])
		),
		layers: schedule.layers.map((layer, index) => ({
			layer,
			name: layerName(schedule.layers.length, index),
			duty: formatDuty(layer),
			days: week.map((day) => segments(schedule, day.from, day.to, index))
		})),
		month,
		handovers: handovers(schedule, now, 14),
		audit: schedule.audit,
		resolver: {
			date,
			time,
			coverage: Number.isNaN(at.getTime()) ? coverageAt(schedule, now) : coverageAt(schedule, at)
		},
		target: { startsAt: target.startsAt, endsAt: target.endsAt, person: target.person },
		feedUrl: `${FEED_ORIGIN}/ical/${params.workspace}/${schedule.id}/${schedule.feedToken}`
	};
};

export const actions: Actions = {
	override: async ({ request, params }) => {
		const form = await request.formData();
		const full = form.get('mode') !== 'partial';

		const startsAt = full
			? String(form.get('targetStart'))
			: `${form.get('startDate')}T${form.get('startTime')}:00Z`;
		const endsAt = full
			? String(form.get('targetEnd'))
			: `${form.get('endDate')}T${form.get('endTime')}:00Z`;

		if (Number.isNaN(Date.parse(startsAt)) || Number.isNaN(Date.parse(endsAt))) {
			return fail(400, { error: 'Give the override a start and an end.' });
		}

		const outcome = addOverride(params.id, {
			person: String(form.get('person')),
			startsAt,
			endsAt,
			reason: String(form.get('reason') ?? '').trim() || 'No reason given'
		});

		if ('error' in outcome) return fail(400, { error: outcome.error });
	},

	duplicate: async ({ params }) => {
		const copy = duplicateSchedule(params.id);
		if (!copy) error(404, `No schedule called ${params.id}.`);
		redirect(303, `/${params.workspace}/on-call/${copy.id}`);
	},

	resume: async ({ params }) => {
		resumeSchedule(params.id);
	},

	archive: async ({ params }) => {
		archiveSchedule(params.id);
		redirect(303, `/${params.workspace}/on-call`);
	}
};
