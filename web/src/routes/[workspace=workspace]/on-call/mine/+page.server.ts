import { formatShift, shiftsFor } from '$lib/oncall';
import { listSchedules, listSwapRequests, requestSwap } from '$lib/server/oncall';
import { getSession } from '$lib/server/session';
import type { Actions, PageServerLoad } from './$types';

const DAY = 86_400_000;

export const load: PageServerLoad = ({ params }) => {
	const now = new Date();
	const me = getSession(params.workspace)?.user.name ?? '';
	const schedules = listSchedules();

	const shifts = shiftsFor(schedules, me, now, 7).map((shift, index) => ({
		id: `${shift.scheduleId}-${index}`,
		when: formatShift(shift, now.getTime()),
		schedule: shift.schedule,
		startsAt: shift.startsAt,
		endsAt: shift.endsAt
	}));

	const start = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), 1));
	const length = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 0)).getUTCDate();
	const mine = shiftsFor(schedules, me, start, length);

	const days = new Set<string>();
	for (const shift of mine) {
		// A shift crossing midnight marks both UTC days
		for (
			let at = Date.parse(shift.startsAt);
			at < Date.parse(shift.endsAt);
			at = Math.floor(at / DAY) * DAY + DAY
		) {
			days.add(new Date(at).toISOString().slice(0, 10));
		}
	}

	return {
		now: now.getTime(),
		me,
		shifts,
		requests: listSwapRequests(),
		month: {
			label: start.toLocaleDateString('en-GB', { month: 'long', year: 'numeric', timeZone: 'UTC' }),
			blanks: (start.getUTCDay() + 6) % 7,
			length,
			today: now.toISOString().slice(0, 10),
			prefix: start.toISOString().slice(0, 8),
			days: [...days]
		}
	};
};

export const actions: Actions = {
	swap: async ({ request }) => {
		const form = await request.formData();
		const person = String(form.get('person'));
		const when = String(form.get('when'));

		if (!when) return;

		requestSwap(`Swap ${when} with ${person}`, String(form.get('message') ?? '').trim());
		return { person };
	}
};
