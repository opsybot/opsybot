import { parseFilters } from '$lib/insights';
import { getOverview, insightsAvailable } from '$lib/server/insights';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => {
	const available = insightsAvailable();
	return {
		available,
		overview: available ? getOverview(parseFilters(url)) : null
	};
};
