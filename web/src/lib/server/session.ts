import type { Cookies } from '@sveltejs/kit';
import { WORKSPACE_COOKIE, type Session, type SessionUser, type Workspace } from '$lib/session';
import { scenario } from './fixtures';

const ORGANIZATION = 'Acme';

const WORKSPACES: Workspace[] = [
	{ id: 'acme-corp', name: 'Acme Corp', environment: 'production', health: 'degraded' },
	{ id: 'acme-labs', name: 'Acme Labs', environment: 'production', health: 'operational' },
	{ id: 'acme-eu', name: 'Acme EU', environment: 'staging', health: 'operational' }
];

function user(): SessionUser {
	const state = scenario();
	const offCall = state === 'not-on-call' || state === 'empty';
	return { name: 'Maya Chen', onCallFor: offCall ? null : 'payments' };
}

export function getSession(cookies: Cookies): Session {
	const id = cookies.get(WORKSPACE_COOKIE);
	const activeWorkspace = WORKSPACES.find((workspace) => workspace.id === id) ?? WORKSPACES[0];

	return {
		organization: ORGANIZATION,
		user: user(),
		workspaces: WORKSPACES,
		activeWorkspace
	};
}
