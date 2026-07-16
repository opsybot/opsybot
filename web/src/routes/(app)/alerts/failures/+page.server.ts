import { listFailures } from '$lib/server/alerts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({
	failures: listFailures()
});
