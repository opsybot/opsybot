export type WorkspaceHealth = 'operational' | 'degraded' | 'outage';

export type Workspace = {
	id: string;
	name: string;
	environment: string;
	health: WorkspaceHealth;
};

export type SessionUser = {
	name: string;
	onCallFor: string | null;
};

export type Session = {
	organization: string;
	user: SessionUser;
	workspaces: Workspace[];
	activeWorkspace: Workspace;
};

export const WORKSPACE_COOKIE = 'opsybot_workspace';
export const WORKSPACE_COOKIE_MAX_AGE = 60 * 60 * 24 * 365;

export function healthColour(health: WorkspaceHealth): string {
	if (health === 'outage') return 'var(--critical)';
	if (health === 'degraded') return 'var(--warning)';
	return 'var(--primary)';
}
