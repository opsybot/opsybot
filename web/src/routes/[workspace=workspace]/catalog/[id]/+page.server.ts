import { error, fail, redirect } from '@sveltejs/kit';
import { listAlerts } from '$lib/server/alerts';
import { getService, serviceActivity, serviceNames } from '$lib/server/catalog';
import { save } from '../save';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url, cookies }) => {
	const service = await getService(cookies, params.workspace, params.id);
	if (!service) error(404, `No service called ${params.id}.`);

	const { alerts } = await listAlerts(cookies, params.workspace, { status: ['open', 'acked'] });

	return {
		service,
		activity: await serviceActivity(cookies, params.workspace, params.id, alerts),
		names: await serviceNames(cookies, params.workspace),
		dialogOpen: url.searchParams.has('edit')
	};
};

export const actions: Actions = {
	save: async ({ request, params, cookies }) => {
		const outcome = await save(cookies, params.workspace, request);
		if ('error' in outcome) return fail(400, outcome);

		if (outcome.slug !== params.id) redirect(303, `/${params.workspace}/catalog/${outcome.slug}`);
		return { saved: outcome.slug };
	}
};
