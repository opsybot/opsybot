import { listAudit } from '$lib/server/admin';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ entries: listAudit() });
