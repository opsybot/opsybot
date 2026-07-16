import { getContext, setContext } from 'svelte';
import { THEME_COOKIE, THEME_COOKIE_MAX_AGE, type Theme } from './theme';

const THEME_KEY = Symbol('theme');

class ThemeState {
	current = $state<Theme>('dark');

	constructor(initial: Theme) {
		this.current = initial;
	}

	toggle() {
		this.current = this.current === 'dark' ? 'light' : 'dark';
		document.documentElement.classList.toggle('dark', this.current === 'dark');
		document.cookie = `${THEME_COOKIE}=${this.current}; path=/; max-age=${THEME_COOKIE_MAX_AGE}; samesite=lax`;
	}
}

export function setTheme(initial: Theme): ThemeState {
	return setContext(THEME_KEY, new ThemeState(initial));
}

export function useTheme(): ThemeState {
	return getContext<ThemeState>(THEME_KEY);
}
