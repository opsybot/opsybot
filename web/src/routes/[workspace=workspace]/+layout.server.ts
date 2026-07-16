import { error } from '@sveltejs/kit';
import { SIDEBAR_COOKIE_NAME } from '$lib/components/ui/sidebar/constants';
import { getNavCounts } from '$lib/server/dashboard';
import { getSession } from '$lib/server/session';
import { WORKSPACE_COOKIE, WORKSPACE_COOKIE_MAX_AGE } from '$lib/session';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = ({ params, cookies }) => {
	const session = getSession(params.workspace);
	if (!session) error(404, 'Workspace not found');

	if (cookies.get(WORKSPACE_COOKIE) !== params.workspace) {
		cookies.set(WORKSPACE_COOKIE, params.workspace, {
			path: '/',
			maxAge: WORKSPACE_COOKIE_MAX_AGE,
			sameSite: 'lax'
		});
	}

	return {
		session,
		counts: getNavCounts(),
		sidebarOpen: cookies.get(SIDEBAR_COOKIE_NAME) !== 'false'
	};
};
