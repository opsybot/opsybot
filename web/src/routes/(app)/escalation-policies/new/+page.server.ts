import { fail } from '@sveltejs/kit';
import { analyzeTree, mkLevel, parseTree, saveBlocked, type Tree } from '$lib/escalation';
import { createPolicy } from '$lib/server/escalation';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	const starter: Tree = {
		name: '',
		team: 'payments',
		repeat: '2',
		nodes: [mkLevel({ targets: [{ type: 'schedule', value: 'payments-primary' }] })]
	};
	return { starter };
};

export const actions: Actions = {
	save: async ({ request }) => {
		const form = await request.formData();
		const parsed = parseTree(String(form.get('tree') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });
		if (!parsed.tree.name.trim()) return fail(400, { error: 'Give the policy a name.' });
		if (saveBlocked(analyzeTree(parsed.tree))) {
			return fail(400, { error: 'Resolve the highlighted issues before saving.' });
		}
		const policy = createPolicy(parsed.tree);
		return { saved: true, id: policy.id };
	}
};
