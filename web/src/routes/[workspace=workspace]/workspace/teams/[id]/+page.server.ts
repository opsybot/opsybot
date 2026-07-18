import { error, fail } from '@sveltejs/kit';
import { parseTeam } from '$lib/admin';
import { archiveTeam, getTeam, listMembers, unarchiveTeam, updateTeam } from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => {
	const team = await getTeam(cookies, params.workspace, params.id);
	if (!team) error(404, `No team called ${params.id}.`);
	return { team, members: await listMembers(cookies, params.workspace) };
};

export const actions: Actions = {
	save: async ({ request, cookies, params }) => {
		const roster = (await listMembers(cookies, params.workspace)).map((member) => member.name);
		const parsed = parseTeam(await request.formData(), roster);
		if ('error' in parsed) return fail(400, { error: parsed.error });
		if (!(await updateTeam(cookies, params.workspace, params.id, parsed.name, parsed.members)))
			return fail(404, { error: 'That team no longer exists.' });
		return { saved: true };
	},
	archive: async ({ cookies, params }) => {
		if (!(await archiveTeam(cookies, params.workspace, params.id)))
			return fail(400, { error: 'Could not archive the team.' });
		return { archived: true };
	},
	unarchive: async ({ cookies, params }) => {
		if (!(await unarchiveTeam(cookies, params.workspace, params.id)))
			return fail(400, { error: 'Could not restore the team.' });
		return { unarchived: true };
	}
};
