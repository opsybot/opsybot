import type { Handle } from '@sveltejs/kit';
import { parseTheme, THEME_COOKIE } from '$lib/theme';
import { api, SESSION_COOKIE } from '$lib/server/api';
import type { AuthUser } from '$lib/session';

type SessionUserDTO = { id: string; name: string; email: string; timezone: string };

export const handle: Handle = async ({ event, resolve }) => {
	const theme = parseTheme(event.cookies.get(THEME_COOKIE));
	event.locals.theme = theme;

	event.locals.user = null;
	if (event.cookies.get(SESSION_COOKIE)) {
		const res = await api.get<SessionUserDTO>('/me', event.cookies);
		if (res.ok && res.data) {
			event.locals.user = { id: res.data.id, name: res.data.name, email: res.data.email } satisfies AuthUser;
		}
	}

	return resolve(event, {
		transformPageChunk: ({ html }) => html.replace('%opsybot.theme%', theme === 'dark' ? 'dark' : '')
	});
};
