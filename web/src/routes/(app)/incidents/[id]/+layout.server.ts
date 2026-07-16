import { error } from '@sveltejs/kit';
import { getIncident, listFollowUps } from '$lib/server/incidents';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = ({ params }) => {
	const incident = getIncident(params.id);
	if (!incident) error(404, `No incident with id ${params.id}.`);

	return {
		now: Date.now(),
		incident,
		followUps: listFollowUps().filter((followUp) => followUp.incidentId === incident.id)
	};
};
