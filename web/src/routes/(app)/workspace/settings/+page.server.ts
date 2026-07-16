import { fail } from '@sveltejs/kit';
import { FIELD_TYPES, parseSettings } from '$lib/admin';
import { addField, getSettings, removeField, saveSettings, setThreshold } from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => getSettings();

export const actions: Actions = {
	setThreshold: async ({ request }) => {
		const threshold = String((await request.formData()).get('threshold') ?? '');
		if (!setThreshold(threshold)) return fail(400, { error: 'Pick a severity from the list.' });
		return { ok: true };
	},
	addField: async ({ request }) => {
		const form = await request.formData();
		const name = String(form.get('name') ?? '').replace(/\s+/g, ' ').trim().slice(0, 60);
		const type = String(form.get('type') ?? 'text');
		if (!name) return fail(400, { error: 'Give the field a name.' });
		addField(name, FIELD_TYPES.includes(type) ? type : 'text');
		return { added: true };
	},
	removeField: async ({ request }) => {
		const id = String((await request.formData()).get('id') ?? '');
		if (!removeField(id)) return fail(404, { error: 'That field no longer exists.' });
		return { removed: true };
	},
	save: async ({ request }) => {
		const parsed = parseSettings(String((await request.formData()).get('settings') ?? ''), getSettings());
		if ('error' in parsed) return fail(400, { error: parsed.error });
		saveSettings(parsed);
		return { saved: true };
	}
};
