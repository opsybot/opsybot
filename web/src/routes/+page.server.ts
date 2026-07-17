import { redirect } from '@sveltejs/kit';
import { lastWorkspace } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, locals }) => {
	if (!locals.user) redirect(307, '/login');

	const workspace = await lastWorkspace(cookies);
	if (!workspace) redirect(307, '/login');

	redirect(307, `/${workspace}`);
};
