import { fail } from '@sveltejs/kit';
import { FORMATS } from '$lib/alertsources';
import { createSource, draftSecret, sanitizeMapping } from '$lib/server/alertsources';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return { formats: FORMATS, secret: draftSecret() };
};

export const actions: Actions = {
	create: async ({ request }) => {
		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		const formatId = String(form.get('format') ?? '');
		if (!name) return fail(400, { error: 'Give the source a name.' });
		if (!FORMATS.some((format) => format.id === formatId)) return fail(400, { error: 'Pick a source format.' });

		const posted = String(form.get('secret') ?? '');
		const secret = /^osk_[a-f0-9]{16}$/.test(posted) ? posted : draftSecret();

		createSource({ name, formatId, mapping: sanitizeMapping(String(form.get('mapping') ?? '')), secret });
		return { created: true, name };
	}
};
