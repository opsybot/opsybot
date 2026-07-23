import { fail } from '@sveltejs/kit';
import { isPlatformId } from '$lib/chat';
import {
	connect,
	disconnect,
	linkIdentity,
	listPlatforms,
	setDefaults,
	startIdentityOAuth,
	startOAuth,
	testConnection
} from '$lib/server/chat';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => ({
	platforms: await listPlatforms(cookies, params.workspace)
});

export const actions: Actions = {
	connect: async ({ request, cookies, params }) => {
		const form = await request.formData();
		const provider = String(form.get('platform') ?? '');
		const externalId = String(form.get('externalId') ?? '').trim();
		if (!isPlatformId(provider)) return fail(400, { error: 'That platform is not available.' });
		const { error } = await connect(cookies, params.workspace, provider, externalId || undefined);
		if (error) return fail(400, { error });
		return { connected: true };
	},
	oauthStart: async ({ request, cookies, params }) => {
		const provider = String((await request.formData()).get('platform') ?? '');
		if (!isPlatformId(provider)) return fail(400, { error: 'That platform is not available.' });
		const { url, error } = await startOAuth(cookies, params.workspace, provider);
		if (error || !url) return fail(400, { error: error ?? 'Could not start the connection.' });
		return { oauthUrl: url };
	},
	disconnect: async ({ request, cookies, params }) => {
		const provider = String((await request.formData()).get('platform') ?? '');
		if (!isPlatformId(provider)) return fail(400, { error: 'That platform is not available.' });
		const { error } = await disconnect(cookies, params.workspace, provider);
		if (error) return fail(404, { error });
		return { disconnected: true };
	},
	saveDefaults: async ({ request, cookies, params }) => {
		const form = await request.formData();
		const provider = String(form.get('platform') ?? '');
		if (!isPlatformId(provider)) return fail(400, { error: 'That platform is not available.' });
		const { error } = await setDefaults(cookies, params.workspace, provider, {
			namingPattern: String(form.get('namingPattern') ?? ''),
			announceChannel: String(form.get('announceChannel') ?? ''),
			archiveOnResolve: form.get('archiveOnResolve') === 'true'
		});
		if (error) return fail(404, { error });
		return { saved: true };
	},
	link: async ({ request, cookies, params }) => {
		const provider = String((await request.formData()).get('platform') ?? '');
		if (!isPlatformId(provider)) return fail(400, { error: 'That platform is not available.' });
		const { handle, verified, error } = await linkIdentity(cookies, params.workspace, provider);
		if (error) return fail(400, { error });
		return { linked: true, handle, verified };
	},
	linkOAuth: async ({ request, cookies, params }) => {
		const provider = String((await request.formData()).get('platform') ?? '');
		if (!isPlatformId(provider)) return fail(400, { error: 'That platform is not available.' });
		const { url, error } = await startIdentityOAuth(cookies, params.workspace, provider);
		if (error || !url) return fail(400, { error: error ?? 'Could not start Slack sign-in.' });
		return { oauthUrl: url };
	},
	test: async ({ request, cookies, params }) => {
		const provider = String((await request.formData()).get('platform') ?? '');
		if (!isPlatformId(provider)) return fail(400, { error: 'That platform is not available.' });
		const { delivered, detail, error } = await testConnection(cookies, params.workspace, provider);
		if (error) return fail(400, { error });
		return { tested: true, delivered, detail };
	}
};
