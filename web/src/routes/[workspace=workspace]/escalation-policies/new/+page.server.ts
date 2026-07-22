import { fail } from '@sveltejs/kit';
import { analyzeTree, mkLevel, parseTree, saveBlocked, type Tree } from '$lib/escalation';
import { createPolicy, getDirectory } from '$lib/server/escalation';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const directory = await getDirectory(cookies, params.workspace);
	const starter: Tree = {
		name: '',
		team: directory.teams[0]?.slug ?? '',
		repeat: '2',
		ack: '0',
		nodes: [mkLevel()]
	};
	return { starter, directory };
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
		const { slug, error } = await createPolicy(cookies, params.workspace, parsed.tree);
		if (error || !slug) return fail(400, { error: error ?? 'Could not save the policy.' });
		return { saved: true, id: slug };
	}
};
