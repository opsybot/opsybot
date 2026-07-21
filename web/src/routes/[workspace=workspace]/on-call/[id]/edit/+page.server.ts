import { error, fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { scheduleSchema } from '$lib/schemas/oncall';
import { editSchedule, formOptions, thisMonday, updateSchedule } from '$lib/server/oncall';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const schedule = await editSchedule(cookies, params.workspace, params.id);
	if (!schedule) error(404, `No schedule called ${params.id}.`);

	const { people, teams } = await formOptions(cookies, params.workspace);

	return {
		id: params.id,
		previewFrom: thisMonday(),
		people,
		teams,
		form: await superValidate(schedule, zod4(scheduleSchema))
	};
};

export const actions: Actions = {
	default: async ({ request, params, cookies }) => {
		const form = await superValidate(request, zod4(scheduleSchema));
		if (!form.valid) return fail(400, { form });

		const result = await updateSchedule(cookies, params.workspace, params.id, form.data);
		if (result.nameError || result.error) {
			form.errors.name = [result.nameError ?? result.error ?? 'Could not save the schedule.'];
			return fail(400, { form });
		}

		redirect(303, `/${params.workspace}/on-call/${result.slug}`);
	}
};
