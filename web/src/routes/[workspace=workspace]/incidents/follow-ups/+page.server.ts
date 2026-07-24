import { fail } from '@sveltejs/kit';
import { listIncidents, listOpenFollowUps, meId, toggleFollowUp } from '$lib/server/incidents-api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url, cookies, params }) => {
	const owner = url.searchParams.get('owner');
	const state = url.searchParams.get('state');
	const now = Date.now();

	const overdue = (followUp: { dueAt: string; done: boolean }) =>
		!followUp.done && !!followUp.dueAt && Date.parse(followUp.dueAt) < now;

	const all = await listOpenFollowUps(cookies, params.workspace);
	const followUps = all.filter((followUp) => {
		if (owner && followUp.owner !== owner) return false;
		if (state === 'done') return followUp.done;
		if (state === 'overdue') return overdue(followUp);
		if (state === 'open') return !followUp.done && !overdue(followUp);
		return true;
	});

	const me = await meId(cookies);
	const { incidents } = await listIncidents(cookies, params.workspace, { limit: 100 }, me);

	return {
		now,
		followUps,
		incidents: incidents.map((incident) => ({
			id: incident.id,
			name: incident.name,
			ref: incident.ref ?? incident.id
		})),
		people: [...new Set(all.map((followUp) => followUp.owner).filter(Boolean))]
	};
};

export const actions: Actions = {
	toggle: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await toggleFollowUp(
			cookies,
			params.workspace,
			String(form.get('incident') ?? ''),
			String(form.get('id') ?? ''),
			String(form.get('done')) === 'true'
		);
		if (result.error) return fail(400, { error: result.error });
	}
};
