import { fail } from '@sveltejs/kit';
import { getPlan, isPlanId } from '$lib/billing';
import { changePlan, getBilling } from '$lib/server/billing';
import { deployment } from '$lib/server/fixtures';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => getBilling();

export const actions: Actions = {
	change: async ({ request }) => {
		if (deployment() !== 'cloud') return fail(404, { error: 'No plans on a self-hosted instance.' });
		const planId = String((await request.formData()).get('plan') ?? '');
		if (!isPlanId(planId) || !changePlan(planId)) return fail(400, { error: 'Pick a plan from the list.' });
		return { changed: getPlan(planId)?.name ?? planId };
	}
};
