import { fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { recoveryCodeSchema, totpSchema } from '$lib/schemas/auth';
import { verifyRecoveryCode, verifyTotp } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const mode = url.searchParams.get('mode') === 'recovery' ? 'recovery' : 'code';

	return {
		mode,
		email: 'maya@acme.dev',
		form:
			mode === 'recovery'
				? await superValidate(zod4(recoveryCodeSchema))
				: await superValidate(zod4(totpSchema))
	};
};

export const actions: Actions = {
	code: async ({ request }) => {
		const form = await superValidate(request, zod4(totpSchema));
		if (!form.valid) return fail(400, { form });
		if (!verifyTotp(form.data.code)) return message(form, 'wrong', { status: 400 });

		redirect(303, '/');
	},
	recovery: async ({ request }) => {
		const form = await superValidate(request, zod4(recoveryCodeSchema));
		if (!form.valid) return fail(400, { form });
		if (!verifyRecoveryCode(form.data.code)) return message(form, 'wrong', { status: 400 });

		redirect(303, '/');
	}
};
