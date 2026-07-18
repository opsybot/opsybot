import { fail } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { profileSchema } from '$lib/schemas/auth';
import { apiClient } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const client = apiClient(cookies);
	const { data: me } = await client.GET('/me');
	const { data: ws } = await client.GET('/workspaces');
	const profile = me ?? { id: '', name: '', email: '', timezone: 'UTC', twoFactorEnabled: false };

	return {
		email: profile.email,
		workspaces: ws?.items ?? [],
		form: await superValidate({ name: profile.name, timezone: profile.timezone }, zod4(profileSchema))
	};
};

export const actions: Actions = {
	save: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(profileSchema));
		if (!form.valid) return fail(400, { form });

		const { error } = await apiClient(cookies).PATCH('/me', {
			body: { name: form.data.name, timezone: form.data.timezone }
		});
		if (error) return message(form, error.detail ?? 'Could not save your profile.', { status: 400 });
		return message(form, 'saved');
	}
};
