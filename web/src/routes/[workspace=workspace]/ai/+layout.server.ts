import { getAiSettings } from '$lib/server/ai';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = () => {
	const { enabled } = getAiSettings();
	return { enabled };
};
