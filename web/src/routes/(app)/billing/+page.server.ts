import { redirect } from '@sveltejs/kit';
import { deployment } from '$lib/server/fixtures';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	redirect(307, deployment() === 'cloud' ? '/billing/plans' : '/billing/license');
};
