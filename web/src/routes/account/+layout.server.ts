import { redirect } from '@sveltejs/kit';
import { apiClient } from '$lib/server/api';
import { WORKSPACE_COOKIE } from '$lib/session';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals, cookies }) => {
	if (!locals.user) redirect(303, '/login');

	const { data } = await apiClient(cookies).GET('/workspaces');
	const workspaces = data?.items ?? [];
	const remembered = cookies.get(WORKSPACE_COOKIE);
	const back = workspaces.find((w) => w.id === remembered) ?? workspaces[0];

	return {
		user: locals.user,
		back: back ? { id: back.id, name: back.name } : null
	};
};
