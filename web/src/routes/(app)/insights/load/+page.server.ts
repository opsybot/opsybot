import { parseFilters } from '$lib/insights';
import { getOnCallLoad, insightsAvailable } from '$lib/server/insights';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => {
	const available = insightsAvailable();
	return {
		available,
		load: available ? getOnCallLoad(parseFilters(url)) : null
	};
};
