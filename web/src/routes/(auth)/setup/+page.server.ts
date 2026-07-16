import { error, fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { setupSchema } from '$lib/schemas/auth';
import { deployment } from '$lib/server/fixtures';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	if (deployment() !== 'self-hosted') error(404, 'This instance is already set up.');
	return { form: await superValidate(zod4(setupSchema)) };
};

export const actions: Actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod4(setupSchema));
		if (!form.valid) return fail(400, { form });
		redirect(303, '/two-factor/setup');
	}
};
