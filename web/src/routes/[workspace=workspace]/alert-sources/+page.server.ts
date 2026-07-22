import { listSources, sourceVolume } from '$lib/server/alertsources';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const [sources, volume] = await Promise.all([
		listSources(cookies, params.workspace),
		sourceVolume(cookies, params.workspace)
	]);
	return {
		sources: sources.map((source) => ({ ...source, volume: volume[source.slug] ?? [] }))
	};
};
