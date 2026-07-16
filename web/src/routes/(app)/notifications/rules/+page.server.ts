import { fail } from '@sveltejs/kit';
import { parseRules } from '$lib/notifications';
import { getRules, saveRules } from '$lib/server/notifications';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => getRules();

export const actions: Actions = {
	save: async ({ request }) => {
		const parsed = parseRules(String((await request.formData()).get('rules') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });
		saveRules(parsed.high, parsed.low, parsed.quietHours);
		return { saved: true };
	}
};
