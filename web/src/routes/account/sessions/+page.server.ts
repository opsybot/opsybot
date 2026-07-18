import { fail } from '@sveltejs/kit';
import { apiClient } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const { data } = await apiClient(cookies).GET('/me/sessions');
	return { sessions: data?.items ?? [] };
};

export const actions: Actions = {
	revoke: async ({ request, cookies }) => {
		const id = String((await request.formData()).get('id') ?? '');
		if (!id) return fail(400, { error: 'Missing session.' });
		const { error } = await apiClient(cookies).DELETE('/me/sessions/{sessionId}', {
			params: { path: { sessionId: id } }
		});
		if (error) return fail(400, { error: 'Could not sign out that session.' });
		return { revoked: true };
	},
	revokeOthers: async ({ cookies }) => {
		const client = apiClient(cookies);
		const { data } = await client.GET('/me/sessions');
		const others = (data?.items ?? []).filter((s) => !s.current);
		for (const s of others)
			await client.DELETE('/me/sessions/{sessionId}', { params: { path: { sessionId: s.id } } });
		return { revokedOthers: others.length };
	}
};
