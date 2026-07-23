import { fail } from '@sveltejs/kit';
import { isChannelType } from '$lib/notifications';
import {
	addChannel,
	confirmVerify,
	listChannels,
	removeChannel,
	startVerify,
	testChannel
} from '$lib/server/notifications';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => ({
	channels: await listChannels(cookies, params.workspace)
});

export const actions: Actions = {
	connect: async ({ request, cookies }) => {
		const form = await request.formData();
		const type = String(form.get('type') ?? '');
		const detail = String(form.get('detail') ?? '').trim();
		const secret = String(form.get('secret') ?? '').trim();
		if (!isChannelType(type)) return fail(400, { error: 'That channel type is not available.' });
		if (!detail) return fail(400, { error: 'Enter the address or URL for this channel.' });
		const added = await addChannel(cookies, { type, detail, secret: secret || undefined });
		if (added.error || !added.channel) return fail(400, { error: added.error ?? 'Could not add that channel.' });
		const started = await startVerify(cookies, added.channel.id);
		if (started.error) return fail(400, { error: started.error });
		return { connected: true, channelId: added.channel.id, method: started.method, detail: started.detail };
	},
	verify: async ({ request, cookies }) => {
		const form = await request.formData();
		const { error } = await confirmVerify(cookies, String(form.get('id') ?? ''), String(form.get('code') ?? ''));
		if (error) return fail(400, { error });
		return { verified: true };
	},
	startVerify: async ({ request, cookies }) => {
		const started = await startVerify(cookies, String((await request.formData()).get('id') ?? ''));
		if (started.error) return fail(400, { error: started.error });
		return { started: true, method: started.method, detail: started.detail };
	},
	test: async ({ request, cookies }) => {
		const result = await testChannel(cookies, String((await request.formData()).get('id') ?? ''));
		if (result.error) return fail(400, { error: result.error });
		return { tested: true, delivered: result.delivered, detail: result.detail };
	},
	remove: async ({ request, cookies }) => {
		const { error } = await removeChannel(cookies, String((await request.formData()).get('id') ?? ''));
		if (error) return fail(404, { error });
		return { removed: true };
	}
};
