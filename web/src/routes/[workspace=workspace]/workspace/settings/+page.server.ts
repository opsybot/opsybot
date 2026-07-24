import { fail } from '@sveltejs/kit';
import { FIELD_TYPES, parseSettings } from '$lib/admin';
import { getSettings, saveSettings, setThreshold } from '$lib/server/admin';
import {
	addField as apiAddField,
	listFields,
	listSeverities,
	removeField as apiRemoveField,
	saveSeverityDefs
} from '$lib/server/incidents-api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => {
	const [settings, severities, fields] = await Promise.all([
		getSettings(cookies, params.workspace),
		listSeverities(cookies, params.workspace),
		listFields(cookies, params.workspace)
	]);
	return {
		...settings,
		severities: severities.map((severity) => ({ id: severity.id, def: severity.def })),
		fields: fields.map((field) => ({
			id: field.id,
			name: field.name,
			type: field.type,
			options: field.options
		}))
	};
};

export const actions: Actions = {
	setThreshold: async ({ request }) => {
		const threshold = String((await request.formData()).get('threshold') ?? '');
		if (!setThreshold(threshold)) return fail(400, { error: 'Pick a severity from the list.' });
		return { ok: true };
	},
	addField: async ({ request, cookies, params }) => {
		const form = await request.formData();
		const name = String(form.get('name') ?? '')
			.replace(/\s+/g, ' ')
			.trim()
			.slice(0, 60);
		const type = String(form.get('type') ?? 'text');
		const options = String(form.get('options') ?? '').trim();
		if (!name) return fail(400, { error: 'Give the field a name.' });
		const result = await apiAddField(cookies, params.workspace, {
			name,
			type: FIELD_TYPES.includes(type) ? type : 'text',
			options: options || undefined
		});
		if (result.error) return fail(400, { error: result.error });
		return { added: true };
	},
	removeField: async ({ request, cookies, params }) => {
		const id = String((await request.formData()).get('id') ?? '');
		const result = await apiRemoveField(cookies, params.workspace, id);
		if (result.error) return fail(404, { error: result.error });
		return { removed: true };
	},
	save: async ({ request, cookies, params }) => {
		const current = await getSettings(cookies, params.workspace);
		const parsed = parseSettings(String((await request.formData()).get('settings') ?? ''), current);
		if ('error' in parsed) return fail(400, { error: parsed.error });
		const severities = await saveSeverityDefs(
			cookies,
			params.workspace,
			parsed.severities.map((severity) => severity.def)
		);
		if (severities.error) return fail(400, { error: severities.error });
		await saveSettings(cookies, params.workspace, parsed);
		return { saved: true };
	}
};
