import { json } from '@sveltejs/kit';
import { apiClient } from '$lib/server/api';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url, cookies }) => {
	const slug = url.searchParams.get('slug') ?? '';
	if (!slug) return json({ checked: false });

	const { data, error } = await apiClient(cookies).GET('/auth/slug-available', {
		params: { query: { slug } }
	});
	if (error || !data) return json({ checked: false });
	return json({ checked: true, ...data });
};
