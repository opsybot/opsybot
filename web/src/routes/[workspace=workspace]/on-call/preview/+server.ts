import { json } from '@sveltejs/kit';
import { previewRows } from '$lib/server/oncall';
import type { Layer } from '$lib/oncall';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, params, cookies }) => {
	const body = (await request.json()) as { timezone?: string; layers?: Layer[]; from?: string };
	const rows = await previewRows(
		cookies,
		params.workspace,
		{ name: '', team: '', timezone: body.timezone ?? 'UTC', layers: body.layers ?? [] },
		body.from ?? new Date().toISOString().slice(0, 10)
	);
	return json(rows);
};
