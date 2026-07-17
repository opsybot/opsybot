import { fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { inviteSchema } from '$lib/schemas/auth';
import { api, cookieValue, SESSION_COOKIE } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

type InvitePreview = { email: string; workspace: string; invitedBy: string; sentAt: string };

export const load: PageServerLoad = async ({ url, cookies }) => {
	const token = url.searchParams.get('token') ?? '';
	const res = await api.post<InvitePreview>('/auth/invite/preview', cookies, { body: { token } });
	const form = await superValidate(zod4(inviteSchema));

	if (res.ok && res.data) {
		return { state: 'valid' as const, invite: res.data, form };
	}
	return {
		state: 'expired' as const,
		invite: { email: '', workspace: 'this workspace', invitedBy: 'an admin', sentAt: '' },
		form
	};
};

export const actions: Actions = {
	default: async ({ request, cookies, url }) => {
		const form = await superValidate(request, zod4(inviteSchema));
		if (!form.valid) return fail(400, { form });

		const token = url.searchParams.get('token') ?? '';
		const res = await api.post('/auth/invite/accept', cookies, {
			body: {
				token,
				name: form.data.name,
				password: form.data.password,
				timezone: form.data.timezone
			}
		});

		if (res.ok) {
			const sessionToken = cookieValue(res.setCookie, SESSION_COOKIE);
			if (sessionToken)
				cookies.set(SESSION_COOKIE, sessionToken, { path: '/', httpOnly: true, sameSite: 'lax' });
			redirect(303, '/');
		}

		const detail =
			res.status === 410
				? 'This invite is no longer valid. Ask an admin to send a new one.'
				: (res.problem?.detail ?? 'Check your details and try again.');
		return message(form, detail, { status: res.status === 410 ? 410 : 400 });
	}
};
