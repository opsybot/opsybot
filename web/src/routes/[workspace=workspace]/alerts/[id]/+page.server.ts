import { error, fail } from '@sveltejs/kit';
import { getAlert, setStatus } from '$lib/server/alerts';
import { escalateAlert } from '$lib/server/escalation';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const alert = await getAlert(cookies, params.workspace, params.id);
	if (!alert) error(404, `No alert with id ${params.id}.`);

	return { now: Date.now(), alert };
};

export const actions: Actions = {
	ack: async ({ params, cookies }) => {
		const outcome = await setStatus(cookies, params.workspace, [params.id], 'acked');
		if (outcome.error) return fail(400, { error: outcome.error });
	},
	resolve: async ({ params, cookies }) => {
		const outcome = await setStatus(cookies, params.workspace, [params.id], 'resolved');
		if (outcome.error) return fail(400, { error: outcome.error });
	},
	escalate: async ({ params, cookies }) => {
		const outcome = await escalateAlert(cookies, params.workspace, params.id);
		if (outcome.error) return fail(400, { error: outcome.error });
	}
};
