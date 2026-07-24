import { fail } from '@sveltejs/kit';
import { listAlerts } from '$lib/server/alerts';
import { getService, listServices, serviceNames } from '$lib/server/catalog';
import { save } from './save';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url, params, cookies }) => {
	const { alerts } = await listAlerts(cookies, params.workspace, { status: ['open', 'acked'] });
	const [services, names] = await Promise.all([
		listServices(cookies, params.workspace, alerts),
		serviceNames(cookies, params.workspace)
	]);

	const editing = url.searchParams.get('edit');
	const service = editing ? await getService(cookies, params.workspace, editing) : null;

	return {
		services,
		names,
		anyService: services.length > 0,
		dialog: {
			open: url.searchParams.has('new') || (!!editing && !!service),
			service: service ?? null
		}
	};
};

export const actions: Actions = {
	save: async ({ request, params, cookies }) => {
		const outcome = await save(cookies, params.workspace, request);
		if ('error' in outcome) return fail(400, outcome);
		return { saved: outcome.slug, created: outcome.created };
	}
};
