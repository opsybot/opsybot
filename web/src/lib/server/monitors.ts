import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import type { Heartbeat, HeartbeatState } from '$lib/alerts';
import { formatDuration } from '$lib/time';
import { apiClient } from './api';

type Schemas = components['schemas'];

function toMonitor(dto: Schemas['AlertMonitor']): Heartbeat {
	return {
		id: dto.id,
		slug: dto.slug,
		name: dto.name,
		state: dto.state as HeartbeatState,
		intervalSeconds: dto.intervalSeconds,
		graceSeconds: dto.graceSeconds,
		interval: formatDuration(dto.intervalSeconds),
		grace: formatDuration(dto.graceSeconds),
		lastSeenAt: dto.lastCheckInAt ?? null,
		dueAt: dto.dueAt ?? null,
		checkInUrl: dto.checkInUrl,
		policy: dto.policySlug
	};
}

export async function listMonitors(cookies: Cookies, workspace: string): Promise<Heartbeat[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/alert-monitors', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map(toMonitor);
}

export async function createMonitor(
	cookies: Cookies,
	workspace: string,
	input: { name: string; intervalSeconds: number; graceSeconds: number; policyRef: string }
): Promise<{ monitor?: Heartbeat; error?: string }> {
	const { data, error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/alert-monitors', {
		params: { path: { workspaceId: workspace } },
		body: {
			name: input.name,
			intervalSeconds: input.intervalSeconds,
			graceSeconds: input.graceSeconds,
			policySlug: input.policyRef
		}
	});
	if (error) return { error: error.detail ?? 'Could not create that monitor.' };
	return { monitor: data ? toMonitor(data) : undefined };
}

export async function updateMonitor(
	cookies: Cookies,
	workspace: string,
	monitorId: string,
	input: { name: string; intervalSeconds: number; graceSeconds: number; policyRef: string }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).PUT(
		'/workspaces/{workspaceId}/alert-monitors/{monitorId}',
		{
			params: { path: { workspaceId: workspace, monitorId } },
			body: {
				name: input.name,
				intervalSeconds: input.intervalSeconds,
				graceSeconds: input.graceSeconds,
				policySlug: input.policyRef
			}
		}
	);
	return error ? { error: error.detail ?? 'Could not save that monitor.' } : {};
}

export async function deleteMonitor(
	cookies: Cookies,
	workspace: string,
	monitorId: string
): Promise<boolean> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/alert-monitors/{monitorId}',
		{ params: { path: { workspaceId: workspace, monitorId } } }
	);
	return !error;
}
