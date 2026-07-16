import { createSilence, endSilence, listSilenceHistory, listSilences } from '$lib/server/alerts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => ({
	now: Date.now(),
	silences: listSilences(),
	history: listSilenceHistory(),
	source: url.searchParams.get('source')
});

export const actions: Actions = {
	create: async ({ request }) => {
		const form = await request.formData();

		const fields = form.getAll('field').map(String);
		const values = form.getAll('value').map(String);

		// A blank condition would match everything, so drop it
		const scope = fields
			.map((field, index) => ({ field, value: values[index]?.trim() ?? '' }))
			.filter((condition) => condition.value)
			.map((condition) =>
				condition.field === 'label'
					? `label ${condition.value.replace(':', ' = ')}`
					: `${condition.field} = ${condition.value}`
			);

		if (!scope.length) return { error: 'Add at least one scope condition.' };

		const startsNow = form.get('start') !== 'later';

		createSilence({
			scope,
			reason: String(form.get('reason') ?? ''),
			startsNow,
			startsAt: startsNow
				? undefined
				: new Date(`${form.get('date')}T${form.get('time')}:00Z`).toISOString(),
			durationHours: Number(String(form.get('duration') ?? '1h').replace('h', ''))
		});
	},

	end: async ({ request }) => {
		const form = await request.formData();
		endSilence(String(form.get('id')));
	}
};
