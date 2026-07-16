import { fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { loginSchema } from '$lib/schemas/auth';
import { attemptLogin } from '$lib/server/auth';
import { deployment } from '$lib/server/fixtures';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => ({
	form: await superValidate(zod4(loginSchema)),
	deployment: deployment()
});

export const actions: Actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod4(loginSchema));
		if (!form.valid) return fail(400, { form });

		const outcome = attemptLogin(form.data.email);

		if (outcome === 'ok') redirect(303, '/two-factor');

		return message(form, outcome, { status: outcome === 'invalid' ? 400 : 403 });
	}
};
