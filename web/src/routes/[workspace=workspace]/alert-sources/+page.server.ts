import { listSources } from '$lib/server/alertsources';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	return { sources: await listSources(cookies, params.workspace) };
};
