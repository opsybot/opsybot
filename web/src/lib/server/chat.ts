import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import type { ChannelDefaults, Connection, Health, Platform, PlatformId } from '$lib/chat';
import { DEFAULT_DEFAULTS, PLATFORMS } from '$lib/chat';
import { apiClient } from './api';

type Schemas = components['schemas'];

function connectionFromApi(dto: Schemas['ChatConnection']): Connection {
	return {
		workspace: dto.externalName,
		health: dto.health as Health,
		healthNote: dto.healthNote,
		defaults: {
			namingPattern: dto.namingPattern || DEFAULT_DEFAULTS.namingPattern,
			announceChannel: dto.announceChannel || DEFAULT_DEFAULTS.announceChannel,
			archiveOnResolve: dto.archiveOnResolve
		},
		linked: dto.linked ?? false,
		linkedHandle: dto.linkedHandle ?? ''
	};
}

export function sanitizeNamingPattern(input: string): string {
	const cleaned = input.replace(/\s+/g, ' ').trim().slice(0, 60);
	return cleaned || DEFAULT_DEFAULTS.namingPattern;
}

export async function listPlatforms(cookies: Cookies, workspace: string): Promise<Platform[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/chat/connections', {
		params: { path: { workspaceId: workspace } }
	});
	const byProvider = new Map<string, Schemas['ChatConnection']>();
	for (const item of data?.items ?? []) byProvider.set(item.provider, item);
	return PLATFORMS.map((platform) => {
		const dto = byProvider.get(platform.id);
		return { ...platform, connection: dto ? connectionFromApi(dto) : null };
	});
}

export async function connect(
	cookies: Cookies,
	workspace: string,
	provider: PlatformId,
	externalId?: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/chat/connections', {
		params: { path: { workspaceId: workspace } },
		body: { provider, externalId: externalId || undefined }
	});
	return error ? { error: error.detail ?? 'Could not connect that provider.' } : {};
}

export async function disconnect(
	cookies: Cookies,
	workspace: string,
	provider: PlatformId
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/chat/connections/{provider}',
		{ params: { path: { workspaceId: workspace, provider } } }
	);
	return error ? { error: error.detail ?? 'Could not disconnect that provider.' } : {};
}

export async function setDefaults(
	cookies: Cookies,
	workspace: string,
	provider: PlatformId,
	defaults: ChannelDefaults
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).PUT(
		'/workspaces/{workspaceId}/chat/connections/{provider}/defaults',
		{
			params: { path: { workspaceId: workspace, provider } },
			body: {
				namingPattern: sanitizeNamingPattern(defaults.namingPattern),
				announceChannel: defaults.announceChannel,
				archiveOnResolve: defaults.archiveOnResolve
			}
		}
	);
	return error ? { error: error.detail ?? 'Could not save those defaults.' } : {};
}

export async function startIdentityOAuth(
	cookies: Cookies,
	workspace: string,
	provider: PlatformId
): Promise<{ url?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/chat/connections/{provider}/identity/start',
		{ params: { path: { workspaceId: workspace, provider } } }
	);
	if (error) return { error: error.detail ?? 'Could not start Slack sign-in.' };
	return { url: data?.authorizeUrl };
}

export async function testConnection(
	cookies: Cookies,
	workspace: string,
	provider: PlatformId
): Promise<{ delivered?: boolean; detail?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/chat/connections/{provider}/test',
		{ params: { path: { workspaceId: workspace, provider } } }
	);
	if (error) return { error: error.detail ?? 'Could not send a test message.' };
	return { delivered: data?.delivered, detail: data?.detail };
}

export async function startOAuth(
	cookies: Cookies,
	workspace: string,
	provider: PlatformId
): Promise<{ url?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/chat/connections/{provider}/oauth/start',
		{ params: { path: { workspaceId: workspace, provider } } }
	);
	if (error) return { error: error.detail ?? 'Could not start the connection.' };
	return { url: data?.authorizeUrl };
}

export async function linkIdentity(
	cookies: Cookies,
	workspace: string,
	provider: PlatformId
): Promise<{ handle?: string; verified?: boolean; error?: string }> {
	const { data, error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/chat/connections/{provider}/link',
		{ params: { path: { workspaceId: workspace, provider } } }
	);
	if (error) return { error: error.detail ?? 'Could not find your account in that workspace.' };
	return { handle: data?.providerHandle, verified: data?.verified };
}
