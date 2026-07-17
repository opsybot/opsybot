import { redirect } from '@sveltejs/kit';
import { api } from '$lib/server/api';
import { WORKSPACE_COOKIE } from '$lib/session';
import type { LayoutServerLoad } from './$types';

type WorkspaceDTO = { id: string; name: string; timezone: string; environment?: string };

export const load: LayoutServerLoad = async ({ locals, cookies }) => {
	if (!locals.user) redirect(303, '/login');

	const res = await api.get<{ items: WorkspaceDTO[] }>('/workspaces', cookies);
	const workspaces = res.data?.items ?? [];
	const remembered = cookies.get(WORKSPACE_COOKIE);
	const back = workspaces.find((w) => w.id === remembered) ?? workspaces[0];

	return {
		user: locals.user,
		back: back ? { id: back.id, name: back.name } : null
	};
};
