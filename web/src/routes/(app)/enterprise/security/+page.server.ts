import { fail } from '@sveltejs/kit';
import { parseSecurityPolicy } from '$lib/enterprise';
import { getSecurity, isLicensed, saveSecurityPolicy } from '$lib/server/enterprise';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => (isLicensed() ? { security: getSecurity() } : {});

export const actions: Actions = {
	save: async ({ request }) => {
		if (!isLicensed()) return fail(403, { error: 'The enterprise license is required.' });
		saveSecurityPolicy(parseSecurityPolicy(await request.formData(), getSecurity()));
		return { saved: true };
	}
};
