import { error, redirect } from '@sveltejs/kit';
import { attachToIncident, escalate, getAlert, setStatus } from '$lib/server/alerts';
import { declareIncident, listIncidents } from '$lib/server/incidents';
import { isActive } from '$lib/incidents';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const alert = getAlert(params.id);
	if (!alert) error(404, `No alert with id ${params.id}.`);

	return {
		now: Date.now(),
		alert,
		incidents: listIncidents()
			.filter(isActive)
			.map((incident) => ({
				id: incident.id,
				name: incident.name,
				severity: incident.severity
			}))
	};
};

export const actions: Actions = {
	ack: async ({ params }) => setStatus(params.id, 'acked'),
	resolve: async ({ params }) => setStatus(params.id, 'resolved'),
	escalate: async ({ params }) => escalate(params.id),

	attach: async ({ request, params }) => {
		const form = await request.formData();
		attachToIncident(params.id, String(form.get('incident')));
	},

	promote: async ({ params }) => {
		const alert = getAlert(params.id);
		if (!alert) error(404);

		const incident = declareIncident({
			name: alert.title,
			severity: alert.severity === 'critical' ? 'SEV1' : alert.severity === 'high' ? 'SEV2' : 'SEV3',
			services: [alert.service],
			lead: 'Maya Chen',
			alerts: [alert.id]
		});

		attachToIncident(alert.id, incident.id);
		setStatus(alert.id, 'acked');

		redirect(303, `/${params.workspace}/incidents/${incident.id}`);
	}
};
