import { fail } from '@sveltejs/kit';
import {
	createWebhook,
	deleteWebhook,
	getDirectory,
	listPolicies,
	listWebhooks
} from '$lib/server/escalation';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const directory = await getDirectory(cookies, params.workspace);
	const [policies, webhooks] = await Promise.all([
		listPolicies(cookies, params.workspace, directory),
		listWebhooks(cookies, params.workspace)
	]);
	return { policies, webhooks };
};

export const actions: Actions = {
	addWebhook: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		const url = String(form.get('url') ?? '').trim();
		const secret = String(form.get('secret') ?? '').trim();
		if (!name || !url) return fail(400, { error: 'A webhook needs a name and a URL.' });

		const { error } = await createWebhook(cookies, params.workspace, { name, url, secret });
		if (error) return fail(400, { error });
		return { webhookSaved: true };
	},

	deleteWebhook: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const { error } = await deleteWebhook(cookies, params.workspace, String(form.get('slug')));
		if (error) return fail(400, { error });
		return { webhookDeleted: true };
	}
};
