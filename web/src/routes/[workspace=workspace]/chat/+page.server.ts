import { fail } from '@sveltejs/kit';
import { isPlatformId } from '$lib/chat';
import {
	linkIdentity,
	listPlatforms,
	startIdentityOAuth,
	startTelegramLink,
	testConnection
} from '$lib/server/chat';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => ({
	platforms: await listPlatforms(cookies, params.workspace)
});

export const actions: Actions = {
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
		if (error || !url) return fail(400, { error: error ?? 'Could not start sign-in.' });
		return { oauthUrl: url };
	},
	linkTelegram: async ({ cookies, params }) => {
		const { url, error } = await startTelegramLink(cookies, params.workspace);
		if (error || !url) return fail(400, { error: error ?? 'Could not start Telegram linking.' });
		return { telegramUrl: url };
	},
	test: async ({ request, cookies, params }) => {
		const provider = String((await request.formData()).get('platform') ?? '');
		if (!isPlatformId(provider)) return fail(400, { error: 'That platform is not available.' });
		const { delivered, detail, error } = await testConnection(cookies, params.workspace, provider);
		if (error) return fail(400, { error });
		return { tested: true, delivered, detail };
	}
};
