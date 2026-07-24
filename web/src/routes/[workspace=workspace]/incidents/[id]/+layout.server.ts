import { error } from '@sveltejs/kit';
import { getIncidentDetail, listMembers } from '$lib/server/incidents-api';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ params, cookies }) => {
	const [detail, people] = await Promise.all([
		getIncidentDetail(cookies, params.workspace, params.id),
		listMembers(cookies, params.workspace)
	]);
	if (!detail) error(404, `No incident with id ${params.id}.`);

	return {
		now: Date.now(),
		incident: detail.incident,
		followUps: detail.followUps,
		people
	};
};
