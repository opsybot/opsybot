import { listPolicies } from '$lib/server/escalation';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return { policies: listPolicies() };
};
