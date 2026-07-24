import { error } from '@sveltejs/kit';
import { openAttachmentContent } from '$lib/server/incidents-api';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ params, cookies }) => {
	const upstream = await openAttachmentContent(
		cookies,
		params.workspace,
		params.id,
		params.attachmentId
	);
	if (!upstream.ok || !upstream.body) {
		error(upstream.status === 404 ? 404 : 502, 'That attachment is not available.');
	}
	const headers = new Headers({
		'content-type': upstream.headers.get('content-type') ?? 'application/octet-stream',
		'cache-control': 'private, max-age=300'
	});
	const length = upstream.headers.get('content-length');
	if (length) headers.set('content-length', length);
	return new Response(upstream.body, { headers });
};
