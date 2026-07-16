import { error, fail } from '@sveltejs/kit';
import { parseTeam } from '$lib/admin';
import { getTeam, listMembers, updateTeam } from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const team = getTeam(params.id);
	if (!team) error(404, `No team called ${params.id}.`);
	return { team, members: listMembers() };
};

export const actions: Actions = {
	save: async ({ request, params }) => {
		const parsed = parseTeam(await request.formData(), listMembers().map((member) => member.name));
		if ('error' in parsed) return fail(400, { error: parsed.error });
		if (!updateTeam(params.id, parsed.name, parsed.members))
			return fail(404, { error: 'That team no longer exists.' });
		return { saved: true };
	}
};
