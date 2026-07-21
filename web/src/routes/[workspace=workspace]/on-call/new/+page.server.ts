import { fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { scheduleSchema } from '$lib/schemas/oncall';
import { createSchedule, formOptions, thisMonday } from '$lib/server/oncall';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => {
	const monday = thisMonday();
	const { people, teams } = await formOptions(cookies, params.workspace);

	return {
		previewFrom: monday,
		people,
		teams,
		form: await superValidate(
			{
				name: '',
				team: teams[0] ?? '',
				timezone: 'UTC',
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
	default: async ({ request, params, cookies }) => {
		const form = await superValidate(request, zod4(scheduleSchema));
		if (!form.valid) return fail(400, { form });

		const result = await createSchedule(cookies, params.workspace, form.data);
		if (result.nameError || result.error) {
			form.errors.name = [result.nameError ?? result.error ?? 'Could not create the schedule.'];
			return fail(400, { form });
		}

		redirect(303, `/${params.workspace}/on-call/${result.slug}`);
	}
};
