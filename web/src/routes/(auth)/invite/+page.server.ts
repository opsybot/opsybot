import { fail, redirect } from '@sveltejs/kit';
import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { inviteSchema } from '$lib/schemas/auth';
import { getInvite, tokenState } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => ({
	state: tokenState(url.searchParams.get('token')),
	invite: getInvite(),
	form: await superValidate(zod4(inviteSchema))
});

export const actions: Actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod4(inviteSchema));
		if (!form.valid) return fail(400, { form });
		redirect(303, '/two-factor/setup');
	}
};
