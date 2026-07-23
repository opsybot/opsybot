import type { Cookies } from '@sveltejs/kit';
import { WORKSPACE_COOKIE, type AuthUser, type Session, type Workspace } from '$lib/session';
import { apiClient } from './api';

async function listWorkspaces(cookies: Cookies): Promise<Workspace[]> {
	const { data } = await apiClient(cookies).GET('/workspaces');
	const items = data?.items ?? [];
	return items.map((workspace) => ({
		id: workspace.id,
		name: workspace.name,
		environment: workspace.environment || 'production',
		health: 'operational',
		role: workspace.role ?? 'member'
	}));
}

export async function getSession(
	cookies: Cookies,
	workspaceSlug: string,
	user: AuthUser
): Promise<Session | null> {
	const workspaces = await listWorkspaces(cookies);
	const activeWorkspace = workspaces.find((workspace) => workspace.id === workspaceSlug);
	if (!activeWorkspace) return null;

	return {
		organization: activeWorkspace.name,
		user: { name: user.name, onCallFor: null },
		workspaces,
		activeWorkspace
	};
}

export async function lastWorkspace(cookies: Cookies): Promise<string> {
	const workspaces = await listWorkspaces(cookies);
	const remembered = cookies.get(WORKSPACE_COOKIE);
	const known = workspaces.find((workspace) => workspace.id === remembered);
	return (known ?? workspaces[0])?.id ?? '';
}
