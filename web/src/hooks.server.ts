import type { Handle } from '@sveltejs/kit';
import { parseTheme, THEME_COOKIE } from '$lib/theme';
import { apiClient, SESSION_COOKIE } from '$lib/server/api';
import type { AuthUser } from '$lib/session';

export const handle: Handle = async ({ event, resolve }) => {
	const theme = parseTheme(event.cookies.get(THEME_COOKIE));
	event.locals.theme = theme;

	event.locals.user = null;
	if (event.cookies.get(SESSION_COOKIE)) {
		const { data } = await apiClient(event.cookies).GET('/me');
		if (data) {
			event.locals.user = { id: data.id, name: data.name, email: data.email } satisfies AuthUser;
		}
	}

	return resolve(event, {
		transformPageChunk: ({ html }) => html.replace('%opsybot.theme%', theme === 'dark' ? 'dark' : '')
	});
};
