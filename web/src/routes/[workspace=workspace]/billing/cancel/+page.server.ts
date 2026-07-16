import { fail } from '@sveltejs/kit';
import { cancelPlan, getBilling } from '$lib/server/billing';
import { deployment } from '$lib/server/fixtures';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => getBilling();

export const actions: Actions = {
	cancel: async ({ request }) => {
		if (deployment() !== 'cloud') return fail(404, { error: 'Nothing to cancel on a self-hosted instance.' });
		const reason = String((await request.formData()).get('reason') ?? '').slice(0, 1000);
		cancelPlan(reason);
		return { cancelled: true };
	}
};
