import { error, fail } from '@sveltejs/kit';
import { analyzeTree, parseTree, saveBlocked } from '$lib/escalation';
import { getPolicy, updateTree } from '$lib/server/escalation';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const policy = getPolicy(params.id);
	if (!policy) error(404, `No escalation policy called ${params.id}.`);
	return { id: policy.id, tree: policy.tree };
};

export const actions: Actions = {
	save: async ({ request, params }) => {
		const form = await request.formData();
		const parsed = parseTree(String(form.get('tree') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });
		if (!parsed.tree.name.trim()) return fail(400, { error: 'Give the policy a name.' });
		if (saveBlocked(analyzeTree(parsed.tree))) {
			return fail(400, { error: 'Resolve the highlighted issues before saving.' });
		}
		if (!updateTree(params.id, parsed.tree)) error(404, 'That policy no longer exists.');
		return { saved: true, id: params.id };
	}
};
