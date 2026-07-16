import { redirect } from '@sveltejs/kit';
import type { ColumnFiltersState } from '@tanstack/table-core';
import type { Severity } from '$lib/dashboard';
import { declareIncident, listIncidents } from '$lib/server/incidents';
import type { Actions, PageServerLoad } from './$types';

function filtersFrom(url: URL): ColumnFiltersState {
	const filters: ColumnFiltersState = [
		{ id: 'preset', value: url.searchParams.get('preset') ?? 'active' },
		{ id: 'range', value: url.searchParams.get('range') ?? '30d' }
	];

	for (const key of ['severity', 'status', 'service', 'team', 'lead'] as const) {
		const value = url.searchParams.get(key);
		if (value) filters.push({ id: key, value });
	}

	const query = url.searchParams.get('q');
	if (query) filters.push({ id: 'search', value: query });

	return filters;
}

export const load: PageServerLoad = ({ url }) => {
	const incidents = listIncidents();

	return {
		now: Date.now(),
		incidents,
		filters: filtersFrom(url),
		openAlerts: incidents.flatMap((incident) =>
			incident.alerts.filter((alert) => alert.status !== 'resolved')
		)
	};
};

export const actions: Actions = {
	declare: async ({ request }) => {
		const form = await request.formData();

		const incident = declareIncident({
			name: String(form.get('name') ?? ''),
			severity: (String(form.get('severity') ?? 'SEV2') as Severity) ?? 'SEV2',
			services: form.getAll('services').map(String),
			lead: String(form.get('lead') ?? 'Maya Chen'),
			alerts: form.getAll('alerts').map(String)
		});

		redirect(303, `/incidents/${incident.id}`);
	}
};
