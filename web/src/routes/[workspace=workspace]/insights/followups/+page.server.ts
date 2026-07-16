import { parseFilters } from '$lib/insights';
import { getFollowupCompletion, insightsAvailable } from '$lib/server/insights';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => {
	const available = insightsAvailable();
	return {
		available,
		followups: available ? getFollowupCompletion(parseFilters(url)) : null
	};
};
