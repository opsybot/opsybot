import { error, fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { setupSchema } from '$lib/schemas/auth';
import { deployment } from '$lib/server/fixtures';
import { api, cookieValue, SESSION_COOKIE } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	if (deployment() !== 'self-hosted') error(404, 'This instance is already set up.');
	return { form: await superValidate(zod4(setupSchema)) };
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(setupSchema));
		if (!form.valid) return fail(400, { form });

		const res = await api.post('/auth/setup', cookies, {
			body: {
				name: form.data.name,
				email: form.data.email,
				password: form.data.password,
				workspace: form.data.workspace,
				timezone: form.data.timezone
			}
		});

		if (res.ok) {
			const token = cookieValue(res.setCookie, SESSION_COOKIE);
			if (token) cookies.set(SESSION_COOKIE, token, { path: '/', httpOnly: true, sameSite: 'lax' });
			redirect(303, '/');
		}

		const detail =
			res.status === 409
				? 'This instance is already set up. Log in instead.'
				: (res.problem?.detail ?? 'Check your details and try again.');
		return message(form, detail, { status: res.status === 409 ? 409 : 400 });
	}
};
