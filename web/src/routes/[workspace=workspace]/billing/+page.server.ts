import { redirect } from '@sveltejs/kit';
import { deployment } from '$lib/server/fixtures';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	redirect(307, deployment() === 'cloud' ? `/${params.workspace}/billing/plans` : `/${params.workspace}/billing/license`);
};
