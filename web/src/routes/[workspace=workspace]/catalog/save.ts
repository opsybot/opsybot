import type { Cookies } from '@sveltejs/kit';
import { createService, updateService } from '$lib/server/catalog';

export async function save(
	cookies: Cookies,
	workspace: string,
	request: Request
): Promise<{ error: string } | { slug: string; created: boolean }> {
	const form = await request.formData();

	const editingSlug = form.get('editing') ? String(form.get('editing')) : null;
	const name = String(form.get('name') ?? '').trim();
	if (!name) return { error: 'Give the service a name.' };

	const input = {
		name,
		team: String(form.get('team') ?? '').trim(),
		description: String(form.get('description') ?? '').trim()
	};

	if (editingSlug) {
		const result = await updateService(cookies, workspace, editingSlug, input);
		if (result.error || !result.slug) {
			return { error: result.error ?? 'Could not update the service.' };
		}
		return { slug: result.slug, created: false };
	}

	const result = await createService(cookies, workspace, input);
	if (result.error || !result.slug) {
		return { error: result.error ?? 'Could not create the service.' };
	}
	return { slug: result.slug, created: true };
}
