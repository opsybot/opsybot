import { listAudit, listMembers } from '$lib/server/admin';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params, url }) => {
	const query = {
		q: url.searchParams.get('q') ?? '',
		actor: url.searchParams.get('actor') ?? '',
		action: url.searchParams.get('action') ?? '',
		cursor: url.searchParams.get('cursor') ?? ''
	};
	const [audit, members] = await Promise.all([
		listAudit(cookies, params.workspace, query),
		listMembers(cookies, params.workspace)
	]);
	return { entries: audit.entries, nextCursor: audit.nextCursor, query, members };
};
