import { fail, redirect, type Cookies } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { recoveryCodeSchema, totpSchema } from '$lib/schemas/auth';
import { api, cookieValue, SESSION_COOKIE } from '$lib/server/api';
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

async function complete(path: string, code: string, cookies: Cookies): Promise<boolean> {
	const res = await api.post(`/auth/two-factor/${path}`, cookies, { body: { code } });
	if (res.ok) {
		const token = cookieValue(res.setCookie, SESSION_COOKIE);
		if (token) cookies.set(SESSION_COOKIE, token, { path: '/', httpOnly: true, sameSite: 'lax' });
	}
	return res.ok;
}

export const actions: Actions = {
	code: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(totpSchema));
		if (!form.valid) return fail(400, { form });
		if (!(await complete('verify', form.data.code, cookies))) return message(form, 'wrong', { status: 400 });

		redirect(303, '/');
	},
	recovery: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(recoveryCodeSchema));
		if (!form.valid) return fail(400, { form });
		if (!(await complete('recovery', form.data.code, cookies))) return message(form, 'wrong', { status: 400 });

		redirect(303, '/');
	}
};
