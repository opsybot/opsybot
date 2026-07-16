import { redirect } from '@sveltejs/kit';
import { lastWorkspace } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ cookies }) => {
	redirect(307, `/${lastWorkspace(cookies)}`);
};
