import { listAudit } from '$lib/server/admin';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => ({
	entries: await listAudit(cookies, params.workspace)
});
