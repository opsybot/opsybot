import { getImportPlan } from '$lib/server/importer';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => getImportPlan();
