import { getBackup } from '$lib/server/operations';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => getBackup();
