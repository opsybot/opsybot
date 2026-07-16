import { getRoles, isLicensed } from '$lib/server/enterprise';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => (isLicensed() ? { roles: getRoles() } : {});
