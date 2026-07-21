import { getContext, setContext } from 'svelte';
import type { Session } from '$lib/session';

const APP_SHELL_KEY = Symbol('app-shell');

class AppShellState {
	commandOpen = $state(false);

	#session: () => Session;

	constructor(session: () => Session) {
		this.#session = session;
	}

	get session(): Session {
		return this.#session();
	}

	openCommand() {
		this.commandOpen = true;
	}

	toggleCommand() {
		this.commandOpen = !this.commandOpen;
	}
}

export function setAppShell(session: () => Session): AppShellState {
	return setContext(APP_SHELL_KEY, new AppShellState(session));
}

export function useAppShell(): AppShellState {
	return getContext<AppShellState>(APP_SHELL_KEY);
}
