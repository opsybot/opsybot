import { fail } from '@sveltejs/kit';
import { parseTeam } from '$lib/admin';
import { createTeam, listMembers, listTeams } from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ teams: listTeams(), members: listMembers() });

export const actions: Actions = {
	create: async ({ request }) => {
		const parsed = parseTeam(await request.formData(), listMembers().map((member) => member.name));
		if ('error' in parsed) return fail(400, { error: parsed.error });
		return { created: createTeam(parsed.name, parsed.members).id };
	}
};
