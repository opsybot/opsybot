import { getDashboard } from '$lib/server/dashboard';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ dashboard: getDashboard() });
