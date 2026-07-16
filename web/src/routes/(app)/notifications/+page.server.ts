import { fail } from '@sveltejs/kit';
import { connectChannel, listChannels, removeChannel } from '$lib/server/notifications';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ channels: listChannels() });

export const actions: Actions = {
	connect: async ({ request }) => {
		const type = String((await request.formData()).get('type') ?? '');
		if (!connectChannel(type)) return fail(400, { error: 'That channel type is not available.' });
		return { connected: true };
	},
	remove: async ({ request }) => {
		const id = String((await request.formData()).get('id') ?? '');
		if (!removeChannel(id)) return fail(404, { error: 'That channel is not connected.' });
		return { removed: true };
	}
};
