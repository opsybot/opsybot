import { formatShift } from '$lib/oncall';
import { formOptions, myShifts } from '$lib/server/oncall';
import type { Actions, PageServerLoad } from './$types';

const DAY = 86_400_000;

export const load: PageServerLoad = async ({ locals, params, cookies }) => {
	const now = new Date();
	const me = locals.user?.name ?? '';

	const week = await myShifts(
		cookies,
		params.workspace,
		now.toISOString(),
		new Date(now.getTime() + 7 * DAY).toISOString()
	);
	const shifts = week.map((shift, index) => ({
		id: `${shift.scheduleSlug}-${index}`,
		when: formatShift(shift, now.getTime()),
		schedule: shift.scheduleSlug,
		startsAt: shift.startsAt,
		endsAt: shift.endsAt
	}));

	const start = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), 1));
	const length = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 0)).getUTCDate();
	const monthShifts = await myShifts(
		cookies,
		params.workspace,
		start.toISOString(),
		new Date(start.getTime() + length * DAY).toISOString()
	);

	const days = new Set<string>();
	for (const shift of monthShifts) {
		// A shift crossing midnight marks both UTC days
		for (
			let at = Date.parse(shift.startsAt);
			at < Date.parse(shift.endsAt);
			at = Math.floor(at / DAY) * DAY + DAY
		) {
			days.add(new Date(at).toISOString().slice(0, 10));
		}
	}

	const { people } = await formOptions(cookies, params.workspace);

	return {
		now: now.getTime(),
		me,
		people,
		shifts,
		requests: [] as { id: string; text: string; message: string; status: 'pending' | 'approved' }[],
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
		return { person };
	}
};
