import { fail } from '@sveltejs/kit';
import { parseProfile } from '$lib/billing';
import { getAccount, saveProfile } from '$lib/server/billing';
import { deployment } from '$lib/server/fixtures';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => getAccount();

export const actions: Actions = {
	saveProfile: async ({ request }) => {
		if (deployment() !== 'cloud') return fail(404, { error: 'No billing account on a self-hosted instance.' });
		saveProfile(parseProfile(await request.formData()));
		return { saved: true };
	}
};
