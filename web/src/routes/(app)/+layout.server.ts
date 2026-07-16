import { SIDEBAR_COOKIE_NAME } from '$lib/components/ui/sidebar/constants';
import { getNavCounts } from '$lib/server/dashboard';
import { getSession } from '$lib/server/session';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = ({ cookies }) => ({
	session: getSession(cookies),
	counts: getNavCounts(),
	sidebarOpen: cookies.get(SIDEBAR_COOKIE_NAME) !== 'false'
});
