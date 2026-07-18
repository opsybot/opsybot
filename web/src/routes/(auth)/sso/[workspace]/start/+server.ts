import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

const BASE = env.OPSYBOT_API_URL ?? 'http://127.0.0.1:8099';

export const GET: RequestHandler = ({ params }) => {
	redirect(302, `${BASE}/v1/auth/sso/${encodeURIComponent(params.workspace)}/start`);
};
