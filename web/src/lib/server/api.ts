import createClient from 'openapi-fetch';
import { env } from '$env/dynamic/private';
import type { Cookies } from '@sveltejs/kit';
import type { paths } from '$lib/api/schema';

const BASE = (env.OPSYBOT_API_URL ?? 'http://127.0.0.1:8099') + '/v1';

export const SESSION_COOKIE = 'opsybot_session';
export const PENDING_COOKIE = 'opsybot_2fa';

export function apiClient(cookies: Cookies) {
	const client = createClient<paths>({ baseUrl: BASE });
	client.use({
		onRequest({ request }) {
			const parts: string[] = [];
			const session = cookies.get(SESSION_COOKIE);
			if (session) parts.push(`${SESSION_COOKIE}=${session}`);
			const pending = cookies.get(PENDING_COOKIE);
			if (pending) parts.push(`${PENDING_COOKIE}=${pending}`);
			if (parts.length) request.headers.set('cookie', parts.join('; '));
			return request;
		}
	});
	return client;
}

export function cookieValue(setCookie: string | null, name: string): string | null {
	if (!setCookie) return null;
	const match = new RegExp(`${name}=([^;]+)`).exec(setCookie);
	return match ? match[1] : null;
}
