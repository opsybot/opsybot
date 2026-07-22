import type { ColumnFiltersState } from '@tanstack/table-core';
import { listAlerts, setStatus } from '$lib/server/alerts';
import type { Actions, PageServerLoad } from './$types';

function filtersFrom(url: URL): ColumnFiltersState {
	const statuses = url.searchParams.getAll('status');
	const filters: ColumnFiltersState = [
		{ id: 'status', value: statuses.length ? statuses : ['open', 'acked'] },
		{ id: 'range', value: url.searchParams.get('range') ?? '24h' }
	];

	for (const key of ['severity', 'source', 'service', 'label'] as const) {
		const value = url.searchParams.get(key);
		if (value) filters.push({ id: key, value });
	}

	const query = url.searchParams.get('q');
	if (query) filters.push({ id: 'search', value: query });

	return filters;
}

export const load: PageServerLoad = async ({ url, params, cookies }) => {
	const alerts = await listAlerts(cookies, params.workspace);

	return {
		now: Date.now(),
		alerts,
		filters: filtersFrom(url),
		sources: [...new Set(alerts.map((alert) => alert.source))],
		services: [...new Set(alerts.map((alert) => alert.service).filter(Boolean))],
		labels: [...new Set(alerts.flatMap((alert) => alert.labels))]
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
