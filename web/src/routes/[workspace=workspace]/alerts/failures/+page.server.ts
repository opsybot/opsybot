import { listFailures } from '$lib/server/alerts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => ({
	failures: await listFailures(cookies, params.workspace)
});
