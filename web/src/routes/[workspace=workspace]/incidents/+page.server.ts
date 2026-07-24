import { fail, redirect } from '@sveltejs/kit';
import type { ColumnFiltersState } from '@tanstack/table-core';
import { declareIncident, listIncidents, listMembers, meId } from '$lib/server/incidents-api';
import { listServices } from '$lib/server/services-api';
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

export const load: PageServerLoad = async ({ url, cookies, params }) => {
	const me = await meId(cookies);
	const [incidentsPage, services, members] = await Promise.all([
		listIncidents(cookies, params.workspace, { limit: 100 }, me),
		listServices(cookies, params.workspace),
		listMembers(cookies, params.workspace)
	]);
	const incidents = incidentsPage.incidents;
	const teams = [...new Set(incidents.map((incident) => incident.team).filter(Boolean))];

	return {
		now: Date.now(),
		incidents,
		filters: filtersFrom(url),
		services: services.map((service) => ({ id: service.id, name: service.name })),
		members,
		filterOptions: {
			services: services.map((service) => service.name),
			teams,
			leads: members.map((member) => member.name)
		}
	};
};

export const actions: Actions = {
	declare: async ({ request, params, cookies }) => {
		const form = await request.formData();

		const result = await declareIncident(cookies, params.workspace, {
			name: String(form.get('name') ?? ''),
			severityLevel: String(form.get('severity') ?? ''),
			serviceIds: form.getAll('services').map(String),
			leadUserId: String(form.get('lead') ?? ''),
			alertIds: form.getAll('alerts').map(String)
		});

		if (result.error || !result.id) {
			return fail(400, { error: result.error ?? 'Could not declare the incident.' });
		}

		redirect(303, `/${params.workspace}/incidents/${result.id}`);
	}
};
