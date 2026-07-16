import { fail } from '@sveltejs/kit';
import { TEMPLATES } from '$lib/workflows';
import { getWorkflow, listWorkflows, retryRun, setEnabled } from '$lib/server/workflows';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return {
		workflows: listWorkflows(),
		templates: TEMPLATES
	};
};

export const actions: Actions = {
	toggle: async ({ request }) => {
		const form = await request.formData();
		const workflow = getWorkflow(String(form.get('id')));
		// Invert the stored state, not a form-supplied value
		if (!workflow || !setEnabled(workflow.id, !workflow.enabled)) {
			return fail(404, { error: 'That workflow no longer exists.' });
		}
		return { toggled: true };
	},

	retry: async ({ request }) => {
		const form = await request.formData();
		if (!retryRun(String(form.get('workflow')), String(form.get('run')))) {
			return fail(400, { error: 'That run cannot be retried.' });
		}
		return { retried: true };
	}
};
