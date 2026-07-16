import { listSources } from '$lib/server/alertsources';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return { sources: listSources() };
};
