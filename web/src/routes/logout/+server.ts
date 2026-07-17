import { redirect } from '@sveltejs/kit';
import { api, PENDING_COOKIE, SESSION_COOKIE } from '$lib/server/api';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ cookies }) => {
	await api.post('/auth/logout', cookies, {});
	cookies.delete(SESSION_COOKIE, { path: '/' });
	cookies.delete(PENDING_COOKIE, { path: '/' });
	redirect(303, '/login');
};
