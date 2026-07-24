import { error } from '@sveltejs/kit';
import { openAttachmentContent } from '$lib/server/incidents-api';
import type { RequestHandler } from './$types';

const FORWARDED = [
	'content-type',
	'content-length',
	'content-disposition',
	'x-content-type-options',
	'content-security-policy'
];

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

	const headers = new Headers({ 'cache-control': 'private, max-age=300' });
	for (const name of FORWARDED) {
		const value = upstream.headers.get(name);
		if (value) headers.set(name, value);
	}
	if (!headers.has('content-type')) headers.set('content-type', 'application/octet-stream');
	if (!headers.has('content-disposition')) headers.set('content-disposition', 'attachment');
	headers.set('x-content-type-options', 'nosniff');

	return new Response(upstream.body, { headers });
};
