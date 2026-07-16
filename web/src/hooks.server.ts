import type { Handle } from '@sveltejs/kit';
import { parseTheme, THEME_COOKIE } from '$lib/theme';

export const handle: Handle = async ({ event, resolve }) => {
	const theme = parseTheme(event.cookies.get(THEME_COOKIE));
	event.locals.theme = theme;

	return resolve(event, {
		transformPageChunk: ({ html }) =>
			html.replace('%opsybot.theme%', theme === 'dark' ? 'dark' : '')
	});
};
