import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import { apiClient } from './api';

type Schemas = components['schemas'];

export type CatalogService = {
	id: string;
	slug: string;
	name: string;
	team: string;
	description: string;
};

function toService(dto: Schemas['Service']): CatalogService {
	return {
		id: dto.id,
		slug: dto.slug,
		name: dto.name,
		team: dto.teamSlug ?? '',
		description: dto.description
	};
}

export async function listTeams(
	cookies: Cookies,
	workspace: string
): Promise<{ slug: string; name: string }[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/teams', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map((team) => ({ slug: team.slug, name: team.name }));
}

export async function listServices(cookies: Cookies, workspace: string): Promise<CatalogService[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/services', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map(toService);
}

export async function createService(
	cookies: Cookies,
	workspace: string,
	input: { name: string; team?: string; description?: string }
): Promise<{ id?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/services', {
		params: { path: { workspaceId: workspace } },
		body: {
			name: input.name,
			teamSlug: input.team || undefined,
			description: input.description || undefined
		}
	});
	if (error || !data) return { error: error?.detail ?? 'Could not create the service.' };
	return { id: data.id };
}

export async function updateService(
	cookies: Cookies,
	workspace: string,
	id: string,
	input: { name: string; team?: string; description?: string }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).PATCH('/workspaces/{workspaceId}/services/{serviceId}', {
		params: { path: { workspaceId: workspace, serviceId: id } },
		body: {
			name: input.name,
			teamSlug: input.team || undefined,
			description: input.description || undefined
		}
	});
	return error ? { error: error.detail ?? 'Could not update the service.' } : {};
}

export async function deleteService(
	cookies: Cookies,
	workspace: string,
	id: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/services/{serviceId}',
		{ params: { path: { workspaceId: workspace, serviceId: id } } }
	);
	return error ? { error: error.detail ?? 'Could not delete the service.' } : {};
}
