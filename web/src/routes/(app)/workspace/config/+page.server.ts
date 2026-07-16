import { fail } from '@sveltejs/kit';
import { applyImport, exportYaml, getDiff } from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ diff: getDiff(), exportYaml: exportYaml() });

export const actions: Actions = {
	apply: async ({ request }) => {
		const decision = String((await request.formData()).get('decision') ?? '');
		if (!applyImport(decision)) return fail(400, { error: 'Resolve the pending decision first.' });
		return { applied: true };
	}
};
