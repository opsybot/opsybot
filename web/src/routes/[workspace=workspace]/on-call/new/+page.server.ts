import { fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { scheduleSchema } from '$lib/schemas/oncall';
import { createSchedule, nameTaken, thisMonday } from '$lib/server/oncall';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const monday = thisMonday();

	return {
		previewFrom: monday,
		form: await superValidate(
			{
				name: '',
				team: 'payments',
				layers: [
					{
						id: 'l-1',
						participants: [],
						rotation: 'weekly' as const,
						intervalDays: 7,
						handoverHour: 9,
						startsOn: monday,
						restrictions: []
					}
				]
			},
			zod4(scheduleSchema),
			{ errors: false }
		)
	};
};

export const actions: Actions = {
	default: async ({ request, params }) => {
		const form = await superValidate(request, zod4(scheduleSchema));
		if (!form.valid) return fail(400, { form });

		if (nameTaken(form.data.name)) {
			form.errors.name = ['A schedule already goes by that name.'];
			return fail(400, { form });
		}

		const schedule = createSchedule(form.data);
		redirect(303, `/${params.workspace}/on-call/${schedule.id}`);
	}
};
