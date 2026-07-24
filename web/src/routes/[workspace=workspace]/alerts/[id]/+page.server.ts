import { error, fail, redirect } from '@sveltejs/kit';
import { getAlert, setStatus } from '$lib/server/alerts';
import { escalateAlert } from '$lib/server/escalation';
import { declareFromAlert, linkAlert, listIncidents } from '$lib/server/incidents-api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const alert = await getAlert(cookies, params.workspace, params.id);
	if (!alert) error(404, `No alert with id ${params.id}.`);

	const { incidents } = await listIncidents(cookies, params.workspace, { active: true, limit: 100 });

	return {
		now: Date.now(),
		alert,
		incidents: incidents.map((incident) => ({
			id: incident.id,
			name: `${incident.ref ?? incident.id} · ${incident.name}`
		}))
	};
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
	},
	declare: async ({ params, cookies }) => {
		const outcome = await declareFromAlert(cookies, params.workspace, { alertId: params.id });
		if (outcome.error || !outcome.id) {
			return fail(400, { error: outcome.error ?? 'Could not declare an incident.' });
		}
		redirect(303, `/${params.workspace}/incidents/${outcome.id}`);
	},
	attach: async ({ request, params, cookies }) => {
		const incidentId = String((await request.formData()).get('incident') ?? '');
		if (!incidentId) return fail(400, { error: 'Pick an incident to attach to.' });
		const outcome = await linkAlert(cookies, params.workspace, incidentId, params.id);
		if (outcome.error) return fail(400, { error: outcome.error });
		redirect(303, `/${params.workspace}/incidents/${incidentId}`);
	}
};

