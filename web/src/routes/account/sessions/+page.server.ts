import { fail } from '@sveltejs/kit';
import { api } from '$lib/server/api';
import type { Actions, PageServerLoad } from './$types';

type SessionDTO = {
	id: string;
	createdAt: string;
	lastSeenAt: string;
	ip?: string;
	userAgent?: string;
	current: boolean;
};

export const load: PageServerLoad = async ({ cookies }) => {
	const res = await api.get<{ items: SessionDTO[] }>('/me/sessions', cookies);
	return { sessions: res.data?.items ?? [] };
};

export const actions: Actions = {
	revoke: async ({ request, cookies }) => {
		const id = String((await request.formData()).get('id') ?? '');
		if (!id) return fail(400, { error: 'Missing session.' });
		const res = await api.del(`/me/sessions/${id}`, cookies);
		if (!res.ok) return fail(400, { error: 'Could not sign out that session.' });
		return { revoked: true };
	},
	revokeOthers: async ({ cookies }) => {
		const list = await api.get<{ items: SessionDTO[] }>('/me/sessions', cookies);
		const others = (list.data?.items ?? []).filter((s) => !s.current);
		for (const s of others) await api.del(`/me/sessions/${s.id}`, cookies);
		return { revokedOthers: others.length };
	}
};
