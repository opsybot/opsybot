import { error, fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { setupSchema } from '$lib/schemas/auth';
import { deployment } from '$lib/server/fixtures';
import { apiClient, cookieValue, SESSION_COOKIE } from '$lib/server/api';
import { setFlash } from '$lib/server/flash';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	if (deployment() !== 'self-hosted') error(404, 'This instance does not host setup.');
	const { data } = await apiClient(cookies).GET('/auth/setup');
	if (data && !data.required) {
		error(404, 'This instance is already set up.');
	}
	return { form: await superValidate(zod4(setupSchema)) };
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(setupSchema));
		if (!form.valid) return fail(400, { form });

		const { error, response } = await apiClient(cookies).POST('/auth/setup', {
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
			setFlash(cookies, { tone: 'success', title: 'Instance ready', message: "You're the admin." });
			redirect(303, '/');
		}

		const detail =
			response.status === 409
				? 'This instance is already set up. Log in instead.'
				: (error.detail ?? 'Check your details and try again.');
		return message(form, detail, { status: response.status === 409 ? 409 : 400 });
	}
};
