import { listAlerts, setStatus, type AlertQuery } from '$lib/server/alerts';
import type { Actions, PageServerLoad } from './$types';

const RANGE_HOURS: Record<string, number> = { '24h': 24, '7d': 168, '30d': 720 };
const PAGE_SIZE = 50;

function queryFrom(url: URL): AlertQuery & { range: string } {
	const statuses = url.searchParams.getAll('status');
	const range = url.searchParams.get('range') ?? '24h';
	const hours = RANGE_HOURS[range] ?? RANGE_HOURS['24h'];

	const filter: AlertQuery & { range: string } = {
		status: statuses.length ? statuses : ['open', 'acked'],
		since: new Date(Date.now() - hours * 3_600_000).toISOString(),
		limit: PAGE_SIZE,
		range
	};

	for (const key of ['severity', 'source', 'service', 'label'] as const) {
		const value = url.searchParams.get(key);
		if (value) filter[key] = [value];
	}

	const search = url.searchParams.get('q');
	if (search) filter.query = search;

	const cursor = url.searchParams.get('cursor');
	if (cursor) filter.cursor = cursor;

	return filter;
}

export const load: PageServerLoad = async ({ url, params, cookies }) => {
	const { range, ...filter } = queryFrom(url);
	const page = await listAlerts(cookies, params.workspace, filter);

	const filtered = ['q', 'severity', 'source', 'service', 'label', 'status'].some((key) =>
		url.searchParams.get(key)
	);

	return {
		now: Date.now(),
		range,
		filtered,
		paged: !!filter.cursor,
		alerts: page.alerts,
		nextCursor: page.nextCursor,
		sources: page.facets.sources,
		services: page.facets.services,
		labels: page.facets.labels
	};
};

export const actions: Actions = {
	bulk: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const status = String(form.get('status'));
		if (status !== 'acked' && status !== 'resolved') return;

		const ids = form.getAll('id').map(String);
		await setStatus(cookies, params.workspace, ids, status);
	}
};
