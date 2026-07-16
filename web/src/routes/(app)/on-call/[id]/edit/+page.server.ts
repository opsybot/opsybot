import { error, fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { scheduleSchema } from '$lib/schemas/oncall';
import { getSchedule, nameTaken, thisMonday, updateSchedule } from '$lib/server/oncall';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const schedule = getSchedule(params.id);
	if (!schedule) error(404, `No schedule called ${params.id}.`);

	return {
		id: schedule.id,
		previewFrom: thisMonday(),
		form: await superValidate(
			{ name: schedule.name, team: schedule.team, layers: schedule.layers },
			zod4(scheduleSchema)
		)
	};
};

export const actions: Actions = {
	default: async ({ request, params }) => {
		const form = await superValidate(request, zod4(scheduleSchema));
		if (!form.valid) return fail(400, { form });

		if (nameTaken(form.data.name, params.id)) {
			form.errors.name = ['A schedule already goes by that name.'];
			return fail(400, { form });
		}

		updateSchedule(params.id, form.data);
		redirect(303, `/on-call/${form.data.name}`);
	}
};
