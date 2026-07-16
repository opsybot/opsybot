import { error } from '@sveltejs/kit';
import { firstBranchKind } from '$lib/escalation';
import { getPolicy } from '$lib/server/escalation';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const policy = getPolicy(params.id);
	if (!policy) error(404, `No escalation policy called ${params.id}.`);
	return {
		id: policy.id,
		tree: policy.tree,
		routing: policy.routing,
		recent: policy.recent,
		branch: firstBranchKind(policy.tree)
	};
};
