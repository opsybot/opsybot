import { incidentActions } from './actions';
import type { Actions } from './$types';

export const actions = incidentActions satisfies Actions;

import { listIncidents } from '$lib/server/incidents';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => ({
	candidates: listIncidents()
		.filter((incident) => incident.id !== params.id)
		.map((incident) => ({ id: incident.id, name: incident.name }))
});
