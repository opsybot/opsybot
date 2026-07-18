import { error, fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { signupSchema } from '$lib/schemas/auth';
import { deployment } from '$lib/server/fixtures';
import { api, cookieValue, SESSION_COOKIE } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	if (deployment() === 'self-hosted') error(404, 'This instance does not accept public sign-ups.');
	return { form: await superValidate(zod4(signupSchema)) };
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(signupSchema));
		if (!form.valid) return fail(400, { form });

		const res = await api.post('/auth/signup', cookies, {
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

		const type = res.problem?.type ?? '';
		const detail = type.endsWith('email-taken')
			? 'That email already has an account. Log in instead.'
			: (res.problem?.detail ?? 'Check your details and try again.');
		return message(form, detail, { status: res.status === 409 ? 409 : 400 });
	}
};
