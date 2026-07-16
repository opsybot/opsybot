import { redirect } from '@sveltejs/kit';
import { deployment } from '$lib/server/fixtures';
import type { LayoutServerLoad } from './$types';

const SELF_ONLY = new Set(['/billing/license', '/billing/delivery']);
const CLOUD_ONLY = new Set(['/billing/plans', '/billing/account', '/billing/cancel']);

export const load: LayoutServerLoad = ({ url }) => {
	const deploy = deployment();
	if (deploy === 'cloud' && SELF_ONLY.has(url.pathname)) redirect(307, '/billing/plans');
	if (deploy === 'self-hosted' && CLOUD_ONLY.has(url.pathname)) redirect(307, '/billing/license');
	return { deployment: deploy };
};
