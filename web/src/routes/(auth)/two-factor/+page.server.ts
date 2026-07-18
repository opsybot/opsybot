import { fail, redirect, type Cookies } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { recoveryCodeSchema, totpSchema } from '$lib/schemas/auth';
import { apiClient, cookieValue, SESSION_COOKIE } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url, locals }) => {
	const mode = url.searchParams.get('mode') === 'recovery' ? 'recovery' : 'code';

	return {
		mode,
		email: locals.user?.email ?? '',
		form:
			mode === 'recovery'
				? await superValidate(zod4(recoveryCodeSchema))
				: await superValidate(zod4(totpSchema))
	};
};

function setSession(cookies: Cookies, response: Response): void {
	const token = cookieValue(response.headers.get('set-cookie'), SESSION_COOKIE);
	if (token) cookies.set(SESSION_COOKIE, token, { path: '/', httpOnly: true, sameSite: 'lax' });
}

export const actions: Actions = {
	code: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(totpSchema));
		if (!form.valid) return fail(400, { form });
		const { error, response } = await apiClient(cookies).POST('/auth/two-factor/verify', {
			body: { code: form.data.code }
		});
		if (error) return message(form, 'wrong', { status: 400 });
		setSession(cookies, response);
		redirect(303, '/');
	},
	recovery: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(recoveryCodeSchema));
		if (!form.valid) return fail(400, { form });
		const { error, response } = await apiClient(cookies).POST('/auth/two-factor/recovery', {
			body: { code: form.data.code }
		});
		if (error) return message(form, 'wrong', { status: 400 });
		setSession(cookies, response);
		redirect(303, '/');
	}
};
