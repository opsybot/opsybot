import { fail } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { profileSchema } from '$lib/schemas/auth';
import { api } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

type Profile = { id: string; name: string; email: string; timezone: string; twoFactorEnabled: boolean };
type WorkspaceDTO = { id: string; name: string; timezone: string; environment?: string };

export const load: PageServerLoad = async ({ cookies }) => {
	const me = await api.get<Profile>('/me', cookies);
	const wsRes = await api.get<{ items: WorkspaceDTO[] }>('/workspaces', cookies);
	const profile = me.data ?? { id: '', name: '', email: '', timezone: 'UTC', twoFactorEnabled: false };

	return {
		email: profile.email,
		workspaces: wsRes.data?.items ?? [],
		form: await superValidate({ name: profile.name, timezone: profile.timezone }, zod4(profileSchema))
	};
};

export const actions: Actions = {
	save: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(profileSchema));
		if (!form.valid) return fail(400, { form });

		const res = await api.patch('/me', cookies, {
			body: { name: form.data.name, timezone: form.data.timezone }
		});
		if (!res.ok) return message(form, res.problem?.detail ?? 'Could not save your profile.', { status: 400 });
		return message(form, 'saved');
	}
};
