import { error, fail } from '@sveltejs/kit';
import { eventsFor, getSource, rotateSecret, sanitizeMapping, saveMapping, setPaused } from '$lib/server/alertsources';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const source = getSource(params.id);
	if (!source) error(404, `No alert source called ${params.id}.`);
	return { source, events: eventsFor(source) };
};

export const actions: Actions = {
	toggle: async ({ params }) => {
		const source = getSource(params.id);
		if (!source || !setPaused(params.id, source.status !== 'paused')) {
			return fail(404, { error: 'That source no longer exists.' });
		}
		return { toggled: true };
	},

	rotate: async ({ params }) => {
		const secret = rotateSecret(params.id);
		if (!secret) return fail(404, { error: 'That source no longer exists.' });
		return { secret };
	},

	saveMapping: async ({ request, params }) => {
		const form = await request.formData();
		const mapping = sanitizeMapping(String(form.get('mapping') ?? ''));
		if (!mapping.length) return fail(400, { error: 'A mapping needs at least one field.' });
		if (!saveMapping(params.id, mapping)) return fail(404, { error: 'That source no longer exists.' });
		return { saved: true };
	}
};
