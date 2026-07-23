import { fail } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { channelSchema } from '$lib/schemas/auth';
import { apiClient } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const { data } = await apiClient(cookies).GET('/me/channels');
	return {
		channels: data?.items ?? [],
		form: await superValidate(zod4(channelSchema))
	};
};

export const actions: Actions = {
	add: async ({ request, cookies }) => {
		const form = await superValidate(request, zod4(channelSchema));
		if (!form.valid) return fail(400, { form });
		const { error, response } = await apiClient(cookies).POST('/me/channels', {
			body: { type: form.data.type, detail: form.data.detail }
		});
		if (!error) return message(form, 'added');
		const detail =
			response.status === 409
				? 'You already added that channel.'
				: (error.detail ?? 'That address or URL is not valid.');
		return message(form, detail, { status: response.status === 409 ? 409 : 400 });
	},
	verify: async ({ request, cookies }) => {
		const data = await request.formData();
		const id = String(data.get('id') ?? '');
		if (!id) return fail(400, { error: 'Missing channel.' });
		const { error, response } = await apiClient(cookies).POST(
			'/me/channels/{channelId}/verify/start',
			{ params: { path: { channelId: id } } }
		);
		if (error) return fail(response.status, { error: 'Could not start verification.' });
		return { started: true };
	},
	remove: async ({ request, cookies }) => {
		const data = await request.formData();
		const id = String(data.get('id') ?? '');
		if (!id) return fail(400, { error: 'Missing channel.' });
		const { error, response } = await apiClient(cookies).DELETE('/me/channels/{channelId}', {
			params: { path: { channelId: id } }
		});
		if (error) return fail(response.status, { error: 'Could not remove that channel.' });
		return { removed: true };
	}
};
