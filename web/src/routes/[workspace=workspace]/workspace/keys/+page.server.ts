import { fail } from '@sveltejs/kit';
import { parseKeyDraft } from '$lib/admin';
import { createKey, listKeys, revokeKey } from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => ({
	keys: await listKeys(cookies, params.workspace)
});

export const actions: Actions = {
	create: async ({ request, cookies, params }) => {
		const draft = parseKeyDraft(await request.formData());
		if ('error' in draft) return fail(400, { error: draft.error });
		const result = await createKey(cookies, params.workspace, draft.name, draft.scopes, draft.kind);
		if (result.error || !result.secret) return fail(400, { error: result.error ?? 'Could not create the key.' });
		return { secret: result.secret, name: draft.name };
	},
	revoke: async ({ request, cookies, params }) => {
		const id = String((await request.formData()).get('id') ?? '');
		if (!(await revokeKey(cookies, params.workspace, id)))
			return fail(404, { error: 'That key no longer exists.' });
		return { revoked: true };
	}
};
