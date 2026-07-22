import { fail, redirect } from '@sveltejs/kit';
import { FORMATS } from '$lib/alertsources';
import { createSource } from '$lib/server/alertsources';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return { formats: FORMATS };
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

		const outcome = await createSource(cookies, params.workspace, { name, formatId });
		if (outcome.error || !outcome.slug) {
			return fail(400, { error: outcome.error ?? 'Could not create the source.' });
		}
		redirect(303, `/${params.workspace}/alert-sources/${outcome.slug}`);
	}
};
