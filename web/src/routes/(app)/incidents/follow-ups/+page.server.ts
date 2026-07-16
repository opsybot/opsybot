import { listFollowUps, listIncidents, toggleFollowUp } from '$lib/server/incidents';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => {
	const owner = url.searchParams.get('owner');
	const state = url.searchParams.get('state');
	const now = Date.now();

	const overdue = (followUp: { dueAt: string; done: boolean }) =>
		!followUp.done && Date.parse(followUp.dueAt) < now;

	const followUps = listFollowUps().filter((followUp) => {
		if (owner && followUp.owner !== owner) return false;
		if (state === 'open') return !followUp.done && !overdue(followUp);
		if (state === 'done') return followUp.done;
		if (state === 'overdue') return overdue(followUp);
		return true;
	});

	return {
		now,
		followUps,
		incidents: listIncidents().map((incident) => ({ id: incident.id, name: incident.name }))
	};
};

export const actions: Actions = {
	toggle: async ({ request }) => {
		const form = await request.formData();
		toggleFollowUp(String(form.get('id')));
	}
};
