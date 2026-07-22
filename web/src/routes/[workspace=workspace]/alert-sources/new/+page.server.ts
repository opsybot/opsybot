import { fail } from '@sveltejs/kit';
import { FORMATS } from '$lib/alertsources';
import { createSource, eventsFor, sanitizeMapping, saveMapping } from '$lib/server/alertsources';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return { formats: FORMATS.filter((format) => format.id !== 'heartbeat') };
};

export const actions: Actions = {
	create: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		const formatId = String(form.get('format') ?? '');
		if (!name) return fail(400, { error: 'Give the source a name.' });
		if (!FORMATS.some((format) => format.id === formatId)) {
			return fail(400, { error: 'Pick a source format.' });
		}

		const { source, error } = await createSource(cookies, params.workspace, { name, formatId });
		if (error || !source) return fail(400, { error: error ?? 'Could not create the source.' });
		return { source };
	},

	mapping: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const slug = String(form.get('slug') ?? '');
		const mapping = sanitizeMapping(String(form.get('mapping') ?? ''));
		if (!mapping.length) return fail(400, { error: 'A mapping needs at least one field.' });
		if (!(await saveMapping(cookies, params.workspace, slug, mapping))) {
			return fail(400, { error: 'Could not save that mapping.' });
		}
		return { saved: true };
	},

	check: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const events = await eventsFor(cookies, params.workspace, String(form.get('slug') ?? ''));
		return { events };
	}
};
