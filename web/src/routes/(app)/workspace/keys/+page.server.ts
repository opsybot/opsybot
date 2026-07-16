import { fail } from '@sveltejs/kit';
import { parseKeyDraft } from '$lib/admin';
import { createKey, listKeys, revokeKey } from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => ({ keys: listKeys() });

export const actions: Actions = {
	create: async ({ request }) => {
		const draft = parseKeyDraft(await request.formData());
		if ('error' in draft) return fail(400, { error: draft.error });
		const { secret } = createKey(draft.name, draft.scopes, draft.kind);
		return { secret, name: draft.name };
	},
	revoke: async ({ request }) => {
		const id = String((await request.formData()).get('id') ?? '');
		if (!revokeKey(id)) return fail(404, { error: 'That key no longer exists.' });
		return { revoked: true };
	}
};
