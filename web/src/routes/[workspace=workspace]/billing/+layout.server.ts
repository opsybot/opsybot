import { redirect } from '@sveltejs/kit';
import { deployment } from '$lib/server/fixtures';
import type { LayoutServerLoad } from './$types';

const SELF_ONLY = new Set(['/billing/license', '/billing/delivery']);
const CLOUD_ONLY = new Set(['/billing/plans', '/billing/account', '/billing/cancel']);

export const load: LayoutServerLoad = ({ params, url }) => {
	const deploy = deployment();
	const path = url.pathname.slice(`/${params.workspace}`.length);
	if (deploy === 'cloud' && SELF_ONLY.has(path)) redirect(307, `/${params.workspace}/billing/plans`);
	if (deploy === 'self-hosted' && CLOUD_ONLY.has(path)) redirect(307, `/${params.workspace}/billing/license`);
	return { deployment: deploy };
};
