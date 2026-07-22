import { error, fail, redirect } from '@sveltejs/kit';
import { listAlerts } from '$lib/server/alerts';
import { getService, serviceActivity, serviceNames } from '$lib/server/catalog';
import { save } from '../save';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url, cookies }) => {
	const service = getService(params.id);
	if (!service) error(404, `No service called ${params.id}.`);

	const alerts = await listAlerts(cookies, params.workspace);

	return {
		service,
		activity: serviceActivity(params.id, alerts),
		names: serviceNames(),
		dialogOpen: url.searchParams.has('edit')
	};
};

export const actions: Actions = {
	save: async ({ request, params }) => {
		const outcome = await save(request);
		if ('error' in outcome) return fail(400, outcome);

		if (outcome.name !== params.id) redirect(303, `/${params.workspace}/catalog/${outcome.name}`);
		return { saved: outcome.name };
	}
};
