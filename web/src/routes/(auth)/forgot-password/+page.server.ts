import { fail } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { forgotPasswordSchema } from '$lib/schemas/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => ({
	form: await superValidate(zod4(forgotPasswordSchema))
});

export const actions: Actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod4(forgotPasswordSchema));
		if (!form.valid) return fail(400, { form });

		// Same reply whether or not the account exists, to avoid account enumeration
		return message(form, 'sent');
	}
};
