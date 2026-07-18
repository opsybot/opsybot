import { error, fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { signupSchema } from '$lib/schemas/auth';
import { deployment } from '$lib/server/fixtures';
import { apiClient, cookieValue, SESSION_COOKIE } from '$lib/server/api';
import { setFlash } from '$lib/server/flash';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	if (deployment() === 'self-hosted') error(404, 'This instance does not accept public sign-ups.');
	return { form: await superValidate(zod4(signupSchema)) };
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(signupSchema));
		if (!form.valid) return fail(400, { form });

		const { error, response } = await apiClient(cookies).POST('/auth/signup', {
			body: {
				name: form.data.name,
				email: form.data.email,
				password: form.data.password,
				workspace: form.data.workspace,
				slug: form.data.slug,
				timezone: form.data.timezone
			}
		});

		if (!error) {
			const token = cookieValue(response.headers.get('set-cookie'), SESSION_COOKIE);
			if (token) cookies.set(SESSION_COOKIE, token, { path: '/', httpOnly: true, sameSite: 'lax' });
			setFlash(cookies, { tone: 'success', title: 'Welcome to Opsybot', message: 'Your workspace is ready.' });
			redirect(303, '/');
		}

		const type = error.type ?? '';
		const detail = type.endsWith('email-taken')
			? 'That email already has an account. Log in instead.'
			: (error.detail ?? 'Check your details and try again.');
		return message(form, detail, { status: response.status === 409 ? 409 : 400 });
	}
};
