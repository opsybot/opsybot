import { fail, redirect } from '@sveltejs/kit';
import { isPlatformId } from '$lib/chat';
import { connect, disconnect, listPlatforms, setDefaults, startOAuth } from '$lib/server/chat';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params, parent }) => {
	const { session } = await parent();
	if (session.activeWorkspace.role !== 'admin') redirect(303, `/${params.workspace}/chat`);
	return { platforms: await listPlatforms(cookies, params.workspace) };
};

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
	}
};
