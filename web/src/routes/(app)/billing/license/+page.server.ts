import { fail } from '@sveltejs/kit';
import { parseLicenseKey } from '$lib/billing';
import { activateLicense, getLicense } from '$lib/server/billing';
import { deployment } from '$lib/server/fixtures';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ license: getLicense() });

export const actions: Actions = {
	activate: async ({ request }) => {
		if (deployment() !== 'self-hosted') return fail(404, { error: 'No license on a cloud instance.' });
		const parsed = parseLicenseKey(await request.formData());
		if ('error' in parsed) return fail(400, { error: parsed.error });
		activateLicense();
		return { activated: true };
	}
};
