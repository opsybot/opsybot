import { fail } from '@sveltejs/kit';
import {
	connect,
	disconnect,
	listPlatforms,
	reconnect,
	setAnnounceChannel,
	setArchiveOnResolve,
	setNamingPattern
} from '$lib/server/chat';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ platforms: listPlatforms() });

export const actions: Actions = {
	connect: async ({ request }) => {
		const id = String((await request.formData()).get('platform') ?? '');
		if (!connect(id)) return fail(400, { error: 'That platform is not available.' });
		return { connected: true };
	},
	disconnect: async ({ request }) => {
		const id = String((await request.formData()).get('platform') ?? '');
		if (!disconnect(id)) return fail(404, { error: 'That platform is not connected.' });
		return { disconnected: true };
	},
	reconnect: async ({ request }) => {
		const id = String((await request.formData()).get('platform') ?? '');
		if (!reconnect(id)) return fail(404, { error: 'That platform is not connected.' });
		return { reconnected: true };
	},
	saveNaming: async ({ request }) => {
		const form = await request.formData();
		const id = String(form.get('platform') ?? '');
		if (!setNamingPattern(id, String(form.get('pattern') ?? '')))
			return fail(404, { error: 'That platform is not connected.' });
		return { saved: true };
	},
	setAnnounce: async ({ request }) => {
		const form = await request.formData();
		const id = String(form.get('platform') ?? '');
		const channel = String(form.get('channel') ?? '');
		if (!setAnnounceChannel(id, channel))
			return fail(400, { error: 'Pick an announcement channel from the list.' });
		return { saved: true };
	},
	setArchive: async ({ request }) => {
		const form = await request.formData();
		const id = String(form.get('platform') ?? '');
		if (!setArchiveOnResolve(id, form.get('archive') === 'true'))
			return fail(404, { error: 'That platform is not connected.' });
		return { saved: true };
	}
};
