import { fail, type Actions } from '@sveltejs/kit';
import type { IncidentStage, RelationKind } from '$lib/incidents';
import {
	addFollowUp,
	changeSeverity,
	moveStatus,
	relateIncident,
	reopenIncident,
	resolveIncident,
	setCustomFields,
	toggleFollowUp,
	unlinkAlert,
	unrelateIncident,
	updateIncident
} from '$lib/server/incidents-api';

type IncidentParams = { workspace: string; id: string };

export const incidentActions = {
	rename: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await updateIncident(cookies, params.workspace, params.id!, {
			name: String(form.get('name') ?? '')
		});
		if (result.error) return fail(400, { error: result.error });
	},

	severity: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await changeSeverity(
			cookies,
			params.workspace,
			params.id!,
			String(form.get('severity') ?? '')
		);
		if (result.error) return fail(400, { error: result.error });
	},

	role: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const role = String(form.get('role'));
		if (role !== 'lead' && role !== 'comms') return fail(400);
		if (role !== 'lead') return;
		const result = await updateIncident(cookies, params.workspace, params.id!, {
			leadUserId: String(form.get('person') ?? '')
		});
		if (result.error) return fail(400, { error: result.error });
	},

	status: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await moveStatus(
			cookies,
			params.workspace,
			params.id!,
			String(form.get('status')) as IncidentStage
		);
		if (result.error) return fail(409, { error: result.error });
	},

	resolve: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const summary = String(form.get('summary') ?? '').trim();
		if (!summary) return fail(400, { error: 'A resolution summary is required.' });
		const result = await resolveIncident(cookies, params.workspace, params.id!, summary);
		if (result.error) return fail(400, { error: result.error });
	},

	reopen: async ({ params, cookies }) => {
		const result = await reopenIncident(cookies, params.workspace, params.id!);
		if (result.error) return fail(409, { error: result.error });
	},

	link: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await relateIncident(
			cookies,
			params.workspace,
			params.id!,
			String(form.get('relation')) as RelationKind,
			String(form.get('incident') ?? '')
		);
		if (result.error) return fail(400, { error: result.error });
	},

	'add-follow-up': async ({ request, params, cookies }) => {
		const form = await request.formData();
		const due = String(form.get('due') ?? '').trim();
		const result = await addFollowUp(cookies, params.workspace, params.id!, {
			title: String(form.get('title') ?? ''),
			ownerUserId: String(form.get('owner') ?? ''),
			dueAt: due ? new Date(due).toISOString() : undefined
		});
		if (result.error) return fail(400, { error: result.error });
	},

	'toggle-follow-up': async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await toggleFollowUp(
			cookies,
			params.workspace,
			params.id!,
			String(form.get('id') ?? ''),
			String(form.get('done')) === 'true'
		);
		if (result.error) return fail(400, { error: result.error });
	},

	'custom-fields': async ({ request, params, cookies }) => {
		const form = await request.formData();
		const fields: Record<string, string> = {};
		for (const [key, value] of form.entries()) {
			if (!key.startsWith('cf:')) continue;
			const id = key.slice(3);
			const next = String(value).trim();
			fields[id] = fields[id] ? `${fields[id]}, ${next}` : next;
		}
		const result = await setCustomFields(cookies, params.workspace, params.id!, fields);
		if (result.error) return fail(400, { error: result.error });
	},

	'unlink-alert': async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await unlinkAlert(
			cookies,
			params.workspace,
			params.id!,
			String(form.get('alert') ?? '')
		);
		if (result.error) return fail(400, { error: result.error });
	},

	unrelate: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const result = await unrelateIncident(
			cookies,
			params.workspace,
			params.id!,
			String(form.get('relation') ?? '')
		);
		if (result.error) return fail(400, { error: result.error });
	},

	'post-update': async () => {},
	entry: async () => {},
	postmortem: async () => {}
} satisfies Actions<IncidentParams>;
