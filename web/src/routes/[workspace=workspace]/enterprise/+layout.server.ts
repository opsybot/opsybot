import { isLicensed } from '$lib/server/enterprise';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = () => ({ licensed: isLicensed() });
