import { fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { loginSchema } from '$lib/schemas/auth';
import { deployment } from '$lib/server/fixtures';
import { api, cookieValue, PENDING_COOKIE, SESSION_COOKIE } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

type LoginResult = {
	status: 'ok' | 'two_factor_required';
	user?: { id: string; name: string; email: string; timezone: string };
};

export const load: PageServerLoad = async () => ({
	form: await superValidate(zod4(loginSchema)),
	deployment: deployment()
});

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(loginSchema));
		if (!form.valid) return fail(400, { form });

		const res = await api.post<LoginResult>('/auth/login', cookies, {
			body: { email: form.data.email, password: form.data.password, remember: form.data.remember }
		});

		if (res.ok && res.data) {
			if (res.data.status === 'two_factor_required') {
				const pending = cookieValue(res.setCookie, PENDING_COOKIE);
				if (pending) cookies.set(PENDING_COOKIE, pending, { path: '/', httpOnly: true, sameSite: 'lax' });
				redirect(303, '/two-factor');
			}
			const token = cookieValue(res.setCookie, SESSION_COOKIE);
			if (token) cookies.set(SESSION_COOKIE, token, { path: '/', httpOnly: true, sameSite: 'lax' });
			redirect(303, '/');
		}

		const type = res.problem?.type ?? '';
		const outcome = type.endsWith('deactivated')
			? 'deactivated'
			: type.endsWith('sso-required')
				? 'sso-required'
				: 'invalid';
		return message(form, outcome, { status: res.status === 403 ? 403 : 400 });
	}
};
