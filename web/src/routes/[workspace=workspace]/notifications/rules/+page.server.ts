import { fail } from '@sveltejs/kit';
import { parseRules } from '$lib/notifications';
import { getRules, saveRules } from '$lib/server/notifications';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ cookies, params }) => getRules(cookies, params.workspace);

export const actions: Actions = {
	save: async ({ request, cookies, params }) => {
		const parsed = parseRules(String((await request.formData()).get('rules') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });
		const { error } = await saveRules(cookies, params.workspace, parsed.high, parsed.low, parsed.quietHours);
		if (error) return fail(400, { error });
		return { saved: true };
	}
};
