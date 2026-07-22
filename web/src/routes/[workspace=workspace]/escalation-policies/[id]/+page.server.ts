import { error, fail, redirect } from '@sveltejs/kit';
import { firstBranchKind } from '$lib/escalation';
import { deletePolicy, getDirectory, getPolicy } from '$lib/server/escalation';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const directory = await getDirectory(cookies, params.workspace);
	const policy = await getPolicy(cookies, params.workspace, params.id, directory);
	if (!policy) error(404, `No escalation policy called ${params.id}.`);
	return {
		id: policy.id,
		tree: policy.tree,
		routing: policy.routing,
		recent: policy.recent,
		routed: policy.routed,
		branch: firstBranchKind(policy.tree)
	};
};

export const actions: Actions = {
	delete: async ({ params, cookies }) => {
		const { error: message } = await deletePolicy(cookies, params.workspace, params.id);
		if (message) return fail(409, { error: message });
		redirect(303, `/${params.workspace}/escalation-policies`);
	}
};
