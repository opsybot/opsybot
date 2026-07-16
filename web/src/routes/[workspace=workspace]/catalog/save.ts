import type { LinkKind } from '$lib/catalog';
import { CATALOG_TEAMS, LINK_KINDS } from '$lib/catalog';
import { createService, getService, nameTaken, updateService } from '$lib/server/catalog';

export async function save(
	request: Request
): Promise<{ error: string } | { name: string; created: boolean }> {
	const form = await request.formData();

	const editingId = form.get('editing') ? String(form.get('editing')) : null;
	const name = String(form.get('name') ?? '').trim();

	if (!name) return { error: 'Give the service a name.' };
	if (!/^[a-z0-9-]+$/.test(name)) {
		return { error: 'Lower case letters, numbers and dashes — it is used in the URL.' };
	}
	if (nameTaken(name, editingId ?? undefined)) {
		return { error: 'A service already goes by that name.' };
	}
	if (editingId && !getService(editingId)) return { error: 'That service no longer exists.' };

	const team = String(form.get('team') ?? '');
	if (!CATALOG_TEAMS.includes(team)) return { error: 'Pick an owning team.' };

	const links = Object.fromEntries(
		LINK_KINDS.map(({ kind }) => [kind, String(form.get(kind) ?? '').trim()])
	) as Record<LinkKind, string>;

	const input = {
		name,
		team,
		description: String(form.get('description') ?? '').trim(),
		links,
		deps: form.getAll('dep').map(String)
	};

	if (editingId) {
		updateService(editingId, input);
		return { name, created: false };
	}

	createService(input);
	return { name, created: true };
}
