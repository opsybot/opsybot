import { getDefinitions } from '$lib/server/insights';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return { definitions: getDefinitions() };
};
