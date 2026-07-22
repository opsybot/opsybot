import { fail } from '@sveltejs/kit';
import { createSilence, endSilence, listSilenceHistory, listSilences } from '$lib/server/alerts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url, params, cookies }) => ({
	now: Date.now(),
	silences: await listSilences(cookies, params.workspace),
	history: await listSilenceHistory(cookies, params.workspace),
	source: url.searchParams.get('source')
});

export const actions: Actions = {
	create: async ({ request, params, cookies }) => {
		const form = await request.formData();

		const fields = form.getAll('field').map(String);
		const values = form.getAll('value').map(String);

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

		const outcome = await createSilence(cookies, params.workspace, {
			scope,
			reason: String(form.get('reason') ?? ''),
			startsNow,
			startsAt: startsNow
				? undefined
				: new Date(`${form.get('date')}T${form.get('time')}:00Z`).toISOString(),
			durationHours: Number(String(form.get('duration') ?? '1h').replace('h', ''))
		});
		if (outcome.error) return fail(400, { error: outcome.error });
	},

	end: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const outcome = await endSilence(cookies, params.workspace, String(form.get('id')));
		if (outcome.error) return fail(400, { error: outcome.error });
	}
};
