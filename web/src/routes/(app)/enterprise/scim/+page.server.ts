import { fail } from '@sveltejs/kit';
import { getScim, isLicensed, rotateScimToken } from '$lib/server/enterprise';
import type { Actions, PageServerLoad } from './$types';

// The SCIM token must never reach an unlicensed client
export const load: PageServerLoad = () => (isLicensed() ? { scim: getScim() } : {});

export const actions: Actions = {
	rotate: async () => {
		if (!isLicensed()) return fail(403, { error: 'The enterprise license is required.' });
		rotateScimToken();
		return { rotated: true };
	}
};
