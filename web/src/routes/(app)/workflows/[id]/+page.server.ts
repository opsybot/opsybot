import { error, fail } from '@sveltejs/kit';
import { getWorkflow, listRoles, parseDefinition, updateWorkflow } from '$lib/server/workflows';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const workflow = getWorkflow(params.id);
	if (!workflow) error(404, `No workflow called ${params.id}.`);

	return {
		initial: {
			name: workflow.name,
			trigger: workflow.trigger,
			conditions: workflow.conditions,
			actions: workflow.actions
		},
		roleNames: listRoles().map((role) => role.name)
	};
};

export const actions: Actions = {
	save: async ({ request, params }) => {
		const form = await request.formData();
		const parsed = parseDefinition(String(form.get('definition') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });

		if (!updateWorkflow(params.id, parsed.definition)) error(404, 'That workflow no longer exists.');
		return { saved: true };
	}
};
