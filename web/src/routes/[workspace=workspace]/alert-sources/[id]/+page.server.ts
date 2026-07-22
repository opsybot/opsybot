import { error, fail } from '@sveltejs/kit';
import {
	eventsFor,
	getSource,
	rotateSecret,
	sanitizeMapping,
	saveMapping,
	setPaused
} from '$lib/server/alertsources';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const source = await getSource(cookies, params.workspace, params.id);
	if (!source) error(404, `No alert source called ${params.id}.`);
	return { source, events: await eventsFor(cookies, params.workspace, params.id) };
};

export const actions: Actions = {
	toggle: async ({ params, cookies }) => {
		const source = await getSource(cookies, params.workspace, params.id);
		if (!source) return fail(404, { error: 'That source no longer exists.' });
		if (!(await setPaused(cookies, params.workspace, params.id, source.status !== 'paused'))) {
			return fail(400, { error: 'Could not change that source.' });
		}
		return { toggled: true };
	},

	rotate: async ({ params, cookies }) => {
		const secret = await rotateSecret(cookies, params.workspace, params.id);
		if (!secret) return fail(404, { error: 'That source no longer exists.' });
		return { secret };
	},

	saveMapping: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const mapping = sanitizeMapping(String(form.get('mapping') ?? ''));
		if (!mapping.length) return fail(400, { error: 'A mapping needs at least one field.' });
		if (!(await saveMapping(cookies, params.workspace, params.id, mapping))) {
			return fail(400, { error: 'Could not save that mapping.' });
		}
		return { saved: true };
	}
};
