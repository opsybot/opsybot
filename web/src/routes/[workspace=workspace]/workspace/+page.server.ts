import { fail } from '@sveltejs/kit';
import { isRole, parseInvite } from '$lib/admin';
import {
	changeRole,
	deactivateMember,
	inviteMember,
	listMembers,
	reactivateMember
} from '$lib/server/admin';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, params }) => ({
	members: await listMembers(cookies, params.workspace)
});

export const actions: Actions = {
	invite: async ({ request, cookies, params }) => {
		const parsed = parseInvite(await request.formData());
		if ('error' in parsed) return fail(400, { error: parsed.error });
		const result = await inviteMember(cookies, params.workspace, parsed.email, parsed.role);
		if (result.error) return fail(400, { error: result.error });
		return { invited: true };
	},
	changeRole: async ({ request, cookies, params }) => {
		const form = await request.formData();
		const role = String(form.get('role') ?? '');
		if (!isRole(role) || !(await changeRole(cookies, params.workspace, String(form.get('id') ?? ''), role)))
			return fail(400, { error: 'Keep at least one admin, and pick a valid role.' });
		return { changed: true };
	},
	deactivate: async ({ request, cookies, params }) => {
		const form = await request.formData();
		const roster = new Set((await listMembers(cookies, params.workspace)).map((member) => member.name));
		const replacements: Record<string, string> = {};
		try {
			const raw = JSON.parse(String(form.get('replacements') ?? '{}'));
			if (raw && typeof raw === 'object')
				for (const [ref, name] of Object.entries(raw as Record<string, unknown>))
					if (typeof name === 'string' && roster.has(name)) replacements[ref] = name;
		} catch {
		}
		if (!(await deactivateMember(cookies, params.workspace, String(form.get('id') ?? ''), replacements)))
			return fail(400, { error: 'Pick a replacement for every reference first.' });
		return { deactivated: true };
	},
	reactivate: async ({ request, cookies, params }) => {
		const id = String((await request.formData()).get('id') ?? '');
		if (!(await reactivateMember(cookies, params.workspace, id)))
			return fail(404, { error: 'That member is not deactivated.' });
		return { reactivated: true };
	}
};
