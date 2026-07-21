import { fail } from '@sveltejs/kit';
import { createHeartbeat, listHeartbeats } from '$lib/server/alerts';
import type { Actions, PageServerLoad } from './$types';

const INGEST_ORIGIN = 'https://in.opsy.bot';

export const load: PageServerLoad = () => ({
	now: Date.now(),
	heartbeats: listHeartbeats()
});

export const actions: Actions = {
	create: async ({ request, params }) => {
		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();

		if (!name) return fail(400, { error: 'Give the monitor a name.' });

		const { slug, token } = createHeartbeat({
			name,
			interval: String(form.get('interval') ?? '5m'),
			grace: String(form.get('grace') ?? '2m'),
			policy: String(form.get('policy') ?? 'platform-default')
		});

		return { url: `${INGEST_ORIGIN}/hb/${params.workspace}/${slug}/${token}` };
	}
};
