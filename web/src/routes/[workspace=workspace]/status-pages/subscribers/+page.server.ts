import {
	listEmails,
	listWebhooks,
	redeliverWebhook,
	removeEmail,
	subscriberCounts
} from '$lib/server/statuspages';
import { formatUtc } from '$lib/time';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => {
	const query = url.searchParams.get('q') ?? '';
	const counts = subscriberCounts();

	return {
		query,
		counts,
		emails: listEmails(query),
		webhooks: listWebhooks().map((webhook) => ({
			url: webhook.url,
			ok: webhook.ok,
			last: `${formatUtc(webhook.lastAt)} · ${webhook.lastResult}`
		}))
	};
};

export const actions: Actions = {
	remove: async ({ request }) => {
		const form = await request.formData();
		removeEmail(String(form.get('address')));
		return { removed: String(form.get('address')) };
	},

	redeliver: async ({ request }) => {
		const form = await request.formData();
		redeliverWebhook(String(form.get('url')));
		return { redelivered: true };
	}
};
