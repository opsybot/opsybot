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

export const load: PageServerLoad = () => ({ members: listMembers() });

export const actions: Actions = {
	invite: async ({ request }) => {
		const parsed = parseInvite(await request.formData());
		if ('error' in parsed) return fail(400, { error: parsed.error });
		const result = inviteMember(parsed.email, parsed.role);
		if ('error' in result) return fail(400, { error: result.error });
		return { invited: true };
	},
	changeRole: async ({ request }) => {
		const form = await request.formData();
		const role = String(form.get('role') ?? '');
		if (!isRole(role) || !changeRole(String(form.get('id') ?? ''), role))
			return fail(400, { error: 'Keep at least one admin, and pick a valid role.' });
		return { changed: true };
	},
	deactivate: async ({ request }) => {
		const form = await request.formData();
		// Client-submitted replacements are filtered against the roster; unknown names are dropped
		const roster = new Set(listMembers().map((member) => member.name));
		const replacements: Record<string, string> = {};
		try {
			const raw = JSON.parse(String(form.get('replacements') ?? '{}'));
			if (raw && typeof raw === 'object')
				for (const [ref, name] of Object.entries(raw as Record<string, unknown>))
					if (typeof name === 'string' && roster.has(name)) replacements[ref] = name;
		} catch {
			// Bad JSON leaves replacements empty and the guard below rejects
		}
		if (!deactivateMember(String(form.get('id') ?? ''), replacements))
			return fail(400, { error: 'Pick a replacement for every reference first.' });
		return { deactivated: true };
	},
	reactivate: async ({ request }) => {
		if (!reactivateMember(String((await request.formData()).get('id') ?? '')))
			return fail(404, { error: 'That member is not deactivated.' });
		return { reactivated: true };
	}
};
