import { fail } from '@sveltejs/kit';
import { ENTRY_TYPES, type EntryType } from '$lib/incidents';
import {
	addAttachment,
	addTimelineEntry,
	editTimelineEntry,
	exportTimeline,
	listEntryRevisions,
	listTimeline,
	removeAttachment,
	uploadImageAttachment
} from '$lib/server/incidents-api';
import { incidentActions } from '../actions';
import type { Actions, PageServerLoad } from './$types';

const TYPES = new Set<string>(ENTRY_TYPES.map((type) => type.id));

function categories(values: string[]): EntryType[] {
	return values.filter((value): value is EntryType => TYPES.has(value));
}

function entryType(value: FormDataEntryValue | null): EntryType {
	const raw = String(value ?? '');
	return TYPES.has(raw) ? (raw as EntryType) : 'observation';
}

export const load: PageServerLoad = async ({ params, cookies, url }) => {
	const filter = categories(url.searchParams.getAll('type'));
	const limit = Number(url.searchParams.get('limit') ?? '') || undefined;
	const page = await listTimeline(cookies, params.workspace, params.id, {
		category: filter,
		limit
	});
	return { filter, entries: page.entries, nextCursor: page.nextCursor };
};

export const actions = {
	...incidentActions,

	addEntry: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const text = String(form.get('text') ?? '').trim();
		if (!text) return fail(400, { error: 'Write what happened before adding the entry.' });
		const at = String(form.get('at') ?? '').trim();
		const result = await addTimelineEntry(cookies, params.workspace, params.id!, {
			text,
			category: entryType(form.get('category')),
			at: at ? new Date(at).toISOString() : undefined,
			idempotencyKey: String(form.get('idempotencyKey') ?? '') || undefined
		});
		if (result.error) return fail(400, { error: result.error });
	},

	editEntry: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const text = String(form.get('text') ?? '').trim();
		if (!text) return fail(400, { error: 'An entry needs text.' });
		const result = await editTimelineEntry(
			cookies,
			params.workspace,
			params.id!,
			String(form.get('entryId') ?? ''),
			{ text, category: entryType(form.get('category')) }
		);
		if (result.error) return fail(400, { error: result.error });
	},

	revisions: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const entryId = String(form.get('entryId') ?? '');
		const revisions = await listEntryRevisions(cookies, params.workspace, params.id!, entryId);
		return { entryId, revisions };
	},

	attach: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const entryId = String(form.get('entryId') ?? '');
		const kind = String(form.get('kind') ?? '');
		const label = String(form.get('label') ?? '').trim();

		if (kind === 'image') {
			const file = form.get('file');
			if (!(file instanceof File) || file.size === 0) {
				return fail(400, { error: 'Choose an image to attach.' });
			}
			const result = await uploadImageAttachment(
				cookies,
				params.workspace,
				params.id!,
				entryId,
				label || file.name,
				file
			);
			if (result.error) return fail(400, { error: result.error });
			return;
		}

		if (kind !== 'link' && kind !== 'log') return fail(400, { error: 'Pick an evidence type.' });
		const url = String(form.get('url') ?? '').trim();
		const body = String(form.get('body') ?? '');
		const result = await addAttachment(cookies, params.workspace, params.id!, entryId, {
			kind,
			label: label || (kind === 'link' ? url : 'Log snippet'),
			url: kind === 'link' ? url : undefined,
			body: kind === 'log' ? body : undefined
		});
		if (result.error) return fail(400, { error: result.error });
	},

	detach: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await removeAttachment(
			cookies,
			params.workspace,
			params.id!,
			String(form.get('attachmentId') ?? '')
		);
		if (result.error) return fail(400, { error: result.error });
	},

	export: async ({ params, cookies }) => {
		const result = await exportTimeline(cookies, params.workspace, params.id!);
		if ('error' in result) return fail(400, { error: result.error });
		return { timelineExport: result };
	}
} satisfies Actions;
