import { getAudit, isLicensed } from '$lib/server/enterprise';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => (isLicensed() ? { audit: getAudit() } : {});
