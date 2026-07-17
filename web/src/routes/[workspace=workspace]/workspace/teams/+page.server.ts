import { fail } from '@sveltejs/kit';
import { parseTeam } from '$lib/admin';
import { createTeam, listMembers, listTeams } from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => ({
	teams: await listTeams(cookies, params.workspace),
	members: await listMembers(cookies, params.workspace)
});

export const actions: Actions = {
	create: async ({ request, cookies, params }) => {
		const roster = (await listMembers(cookies, params.workspace)).map((member) => member.name);
		const parsed = parseTeam(await request.formData(), roster);
		if ('error' in parsed) return fail(400, { error: parsed.error });
		const result = await createTeam(cookies, params.workspace, parsed.name, parsed.members);
		if (result.error) return fail(400, { error: result.error });
		return { created: result.id };
	}
};
