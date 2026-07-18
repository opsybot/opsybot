import { redirect } from '@sveltejs/kit';
import { apiClient, PENDING_COOKIE, SESSION_COOKIE } from '$lib/server/api';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ cookies }) => {
	await apiClient(cookies).POST('/auth/logout');
	cookies.delete(SESSION_COOKIE, { path: '/' });
	cookies.delete(PENDING_COOKIE, { path: '/' });
	redirect(303, '/login');
};
