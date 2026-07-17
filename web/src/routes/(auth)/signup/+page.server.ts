import { error, fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { signupAccountSchema, workspaceSchema } from '$lib/schemas/auth';
import { deployment } from '$lib/server/fixtures';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	if (deployment() === 'self-hosted') error(404, 'This instance does not accept public sign-ups.');

	if (url.searchParams.get('step') === 'workspace') {
		return { step: 'workspace' as const, form: await superValidate(zod4(workspaceSchema)) };
	}

	return { step: 'account' as const, form: await superValidate(zod4(signupAccountSchema)) };
};

export const actions: Actions = {
	account: async ({ request }) => {
		const form = await superValidate(request, zod4(signupAccountSchema));
		if (!form.valid) return fail(400, { form });
		redirect(303, '/signup?step=workspace');
	},
	workspace: async ({ request }) => {
		const form = await superValidate(request, zod4(workspaceSchema));
		if (!form.valid) return fail(400, { form });
		redirect(303, '/login');
	}
};
