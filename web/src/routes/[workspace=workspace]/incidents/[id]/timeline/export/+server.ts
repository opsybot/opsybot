import { error } from '@sveltejs/kit';
import { exportTimeline } from '$lib/server/incidents-api';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ params, cookies, url }) => {
	const result = await exportTimeline(cookies, params.workspace, params.id);
	if ('error' in result) error(404, result.error);

	const text = url.searchParams.get('format') === 'text';
	const name = `incident-${params.id}-timeline.${text ? 'txt' : 'json'}`;
	return new Response(text ? result.text : result.json, {
		headers: {
			'content-type': text ? 'text/plain; charset=utf-8' : 'application/json; charset=utf-8',
			'content-disposition': `attachment; filename="${name}"`,
			'cache-control': 'no-store'
		}
	});
};
