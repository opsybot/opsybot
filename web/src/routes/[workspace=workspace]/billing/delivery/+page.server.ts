import { fail } from '@sveltejs/kit';
import { isChannelId } from '$lib/billing';
import { getDelivery, setChannel, setDeliveryLinked } from '$lib/server/billing';
import { deployment } from '$lib/server/fixtures';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => getDelivery();

export const actions: Actions = {
	connect: async () => {
		if (deployment() !== 'self-hosted') return fail(404, { error: 'No delivery bridge on a cloud instance.' });
		setDeliveryLinked(true);
		return { linked: true };
	},
	disconnect: async () => {
		if (deployment() !== 'self-hosted') return fail(404, { error: 'No delivery bridge on a cloud instance.' });
		setDeliveryLinked(false);
		return { linked: false };
	},
	toggle: async ({ request }) => {
		if (deployment() !== 'self-hosted') return fail(404, { error: 'No delivery bridge on a cloud instance.' });
		const form = await request.formData();
		const id = String(form.get('id') ?? '');
		const on = String(form.get('on') ?? '') === 'true';
		if (!isChannelId(id) || !setChannel(id, on)) return fail(400, { error: 'Unknown channel.' });
		return { toggled: true };
	}
};
