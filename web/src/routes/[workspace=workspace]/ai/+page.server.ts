import { fail } from '@sveltejs/kit';
import { parseModelDraft } from '$lib/ai';
import {
	addModel,
	assignFeature,
	getAiSettings,
	removeModel,
	setDefault,
	setEnabled
} from '$lib/server/ai';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	const { models, defaultModelId, assignments } = getAiSettings();
	return { models, defaultModelId, assignments };
};

export const actions: Actions = {
	toggle: async ({ request }) => {
		const on = (await request.formData()).get('enabled') === 'true';
		if (!setEnabled(on)) return fail(400, { error: 'Connect a model before turning AI on.' });
		return { enabled: on };
	},
	addModel: async ({ request }) => {
		const draft = parseModelDraft(await request.formData());
		if ('error' in draft) return fail(400, { error: draft.error });
		return { added: addModel(draft).name };
	},
	makeDefault: async ({ request }) => {
		if (!setDefault(String((await request.formData()).get('id') ?? '')))
			return fail(404, { error: 'That model is not configured.' });
		return { ok: true };
	},
	assign: async ({ request }) => {
		const form = await request.formData();
		if (!assignFeature(String(form.get('feature') ?? ''), String(form.get('model') ?? '')))
			return fail(400, { error: 'Pick a model from the list.' });
		return { ok: true };
	},
	remove: async ({ request }) => {
		if (!removeModel(String((await request.formData()).get('id') ?? '')))
			return fail(404, { error: 'That model is not configured.' });
		return { removed: true };
	}
};
