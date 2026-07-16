import { fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { resetPasswordSchema } from '$lib/schemas/auth';
import { tokenState } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => ({
	state: tokenState(url.searchParams.get('token')),
	email: 'maya@acme.dev',
	issuedAt: '2026-07-11T07:40:00Z',
	form: await superValidate(zod4(resetPasswordSchema))
});

export const actions: Actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod4(resetPasswordSchema));
		if (!form.valid) return fail(400, { form });
		redirect(303, '/login');
	}
};
