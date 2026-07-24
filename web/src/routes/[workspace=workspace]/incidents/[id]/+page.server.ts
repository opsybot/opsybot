import { listFields, listIncidents, meId } from '$lib/server/incidents-api';
import { incidentActions } from './actions';
import type { Actions, PageServerLoad } from './$types';

export const actions = incidentActions satisfies Actions;

export const load: PageServerLoad = async ({ params, cookies }) => {
	const me = await meId(cookies);
	const [{ incidents }, fieldDefs] = await Promise.all([
		listIncidents(cookies, params.workspace, { limit: 100 }, me),
		listFields(cookies, params.workspace)
	]);
	return {
		candidates: incidents
			.filter((incident) => incident.id !== params.id)
			.map((incident) => ({ id: incident.id, name: `${incident.ref ?? incident.id} · ${incident.name}` })),
		fieldDefs
	};
};
