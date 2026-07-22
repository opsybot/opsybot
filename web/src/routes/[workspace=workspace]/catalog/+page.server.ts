import { fail } from '@sveltejs/kit';
import { listAlerts } from '$lib/server/alerts';
import { getService, listServices, serviceNames } from '$lib/server/catalog';
import { save } from './save';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url, params, cookies }) => {
	const { alerts } = await listAlerts(cookies, params.workspace, { status: ['open', 'acked'] });
	const services = listServices(alerts);

	const editing = url.searchParams.get('edit');
	const service = editing ? getService(editing) : null;

	return {
		services,
		names: serviceNames(),
		anyService: services.length > 0,
		dialog: {
			open: url.searchParams.has('new') || (!!editing && !!service),
			service: service ?? null
		}
	};
};

export const actions: Actions = {
	save: async ({ request }) => {
		const outcome = await save(request);
		if ('error' in outcome) return fail(400, outcome);
		return { saved: outcome.name, created: outcome.created };
	}
};
