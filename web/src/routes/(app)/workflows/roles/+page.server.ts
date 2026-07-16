import { fail } from '@sveltejs/kit';
import { addRole, listRoles, removeRole, updateRole } from '$lib/server/workflows';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return { roles: listRoles() };
};

export const actions: Actions = {
	save: async ({ request }) => {
		const form = await request.formData();
		const id = String(form.get('id') ?? '').trim();
		const name = String(form.get('name') ?? '').trim();
		const description = String(form.get('description') ?? '').trim();

		if (!name || !description) {
			return fail(400, { error: 'A role needs a name and a responsibility description.' });
		}

		if (id) {
			if (!updateRole(id, { name, description })) return fail(404, { error: 'That role no longer exists.' });
			return { saved: true };
		}

		addRole({ name, description });
		return { saved: true };
	},

	remove: async ({ request }) => {
		const form = await request.formData();
		if (!removeRole(String(form.get('id')))) {
			return fail(400, { error: 'That role cannot be removed.' });
		}
		return { removed: true };
	}
};
