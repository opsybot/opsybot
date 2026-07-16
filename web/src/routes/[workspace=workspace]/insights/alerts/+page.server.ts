import { parseFilters } from '$lib/insights';
import { getAlertAnalytics, insightsAvailable } from '$lib/server/insights';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => {
	const available = insightsAvailable();
	return {
		available,
		alerts: available ? getAlertAnalytics(parseFilters(url)) : null
	};
};
