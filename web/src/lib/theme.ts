export type Theme = 'dark' | 'light';

export const DEFAULT_THEME: Theme = 'dark';

export const THEME_COOKIE = 'opsybot_theme';
export const THEME_COOKIE_MAX_AGE = 60 * 60 * 24 * 365;

export function parseTheme(value: string | undefined): Theme {
	return value === 'light' || value === 'dark' ? value : DEFAULT_THEME;
}
