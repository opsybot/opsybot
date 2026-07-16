import { fail } from '@sveltejs/kit';
import { allComponentNames, listMaintenance, scheduleMaintenance } from '$lib/server/statuspages';
import { NOTICES } from '$lib/statuspages';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	const { upcoming, past } = listMaintenance();

	return {
		upcoming,
		past,
		components: allComponentNames(),
		notices: NOTICES
	};
};

export const actions: Actions = {
	schedule: async ({ request }) => {
		const form = await request.formData();

		const title = String(form.get('title') ?? '').trim();
		const components = form.getAll('component').map(String);

		if (!title) return fail(400, { error: 'Give the window a title.' });
		if (!components.length) return fail(400, { error: 'Pick at least one component.' });

		const date = String(form.get('date'));
		const startsAt = `${date}T${form.get('startTime')}:00Z`;
		const endsAt = `${date}T${form.get('endTime')}:00Z`;
		if (Number.isNaN(Date.parse(startsAt)) || Number.isNaN(Date.parse(endsAt))) {
			return fail(400, { error: 'Give the window a date and times.' });
		}

		scheduleMaintenance({
			title,
			description: String(form.get('description') ?? '').trim(),
			components,
			startsAt,
			endsAt,
			notice: String(form.get('notice') ?? NOTICES[1])
		});

		return { scheduled: true };
	}
};
