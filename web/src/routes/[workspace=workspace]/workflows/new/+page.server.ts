import { fail } from '@sveltejs/kit';
import { getTemplate } from '$lib/workflows';
import { createWorkflow, listRoles, parseDefinition } from '$lib/server/workflows';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => {
	const template = getTemplate(url.searchParams.get('template') ?? '');

	return {
		initial: template
			? {
					name: template.name,
					trigger: template.trigger,
					conditions: template.conditions,
					actions: template.actions.map((action, index) => ({
						id: `t${index}`,
						type: action.type,
						config: action.config
					}))
				}
			: null,
		fromTemplate: !!template,
		roleNames: listRoles().map((role) => role.name)
	};
};

export const actions: Actions = {
	save: async ({ request }) => {
		const form = await request.formData();
		const parsed = parseDefinition(String(form.get('definition') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });

		createWorkflow(parsed.definition);
		return { saved: true };
	}
};
