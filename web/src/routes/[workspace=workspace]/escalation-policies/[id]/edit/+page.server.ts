import { error, fail } from '@sveltejs/kit';
import { analyzeTree, parseTree, saveBlocked } from '$lib/escalation';
import { getDirectory, getPolicy, updatePolicy } from '$lib/server/escalation';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const directory = await getDirectory(cookies, params.workspace);
	const policy = await getPolicy(cookies, params.workspace, params.id, directory);
	if (!policy) error(404, `No escalation policy called ${params.id}.`);
	return { id: policy.id, tree: policy.tree, directory };
};

export const actions: Actions = {
	save: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const parsed = parseTree(String(form.get('tree') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });
		if (!parsed.tree.name.trim()) return fail(400, { error: 'Give the policy a name.' });
		if (saveBlocked(analyzeTree(parsed.tree))) {
			return fail(400, { error: 'Resolve the highlighted issues before saving.' });
		}
		const { error: message } = await updatePolicy(cookies, params.workspace, params.id, parsed.tree);
		if (message) return fail(400, { error: message });
		return { saved: true, id: params.id };
	}
};
