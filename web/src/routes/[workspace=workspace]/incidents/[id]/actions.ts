import { fail, type Actions } from '@sveltejs/kit';
import type { Severity } from '$lib/dashboard';
import type { EntryType, IncidentStage, RelationKind } from '$lib/incidents';
import {
	addEntry,
	addFollowUp,
	assignRole,
	changeSeverity,
	linkIncident,
	moveStatus,
	postUpdate,
	rename,
	reopenIncident,
	resolveIncident,
	toggleFollowUp
} from '$lib/server/incidents';
import { advance } from '$lib/server/postmortems';

export const incidentActions = {
	rename: async ({ request, params }) => {
		const form = await request.formData();
		rename(params.id!, String(form.get('name') ?? ''));
	},

	severity: async ({ request, params }) => {
		const form = await request.formData();
		changeSeverity(params.id!, String(form.get('severity')) as Severity);
	},

	role: async ({ request, params }) => {
		const form = await request.formData();
		const role = String(form.get('role'));
		if (role !== 'lead' && role !== 'comms') return fail(400);
		assignRole(params.id!, role, String(form.get('person')));
	},

	status: async ({ request, params }) => {
		const form = await request.formData();
		moveStatus(params.id!, String(form.get('status')) as IncidentStage);
	},

	'post-update': async ({ params }) => {
		postUpdate(params.id!);
	},

	resolve: async ({ request, params }) => {
		const form = await request.formData();
		const summary = String(form.get('summary') ?? '').trim();
		if (!summary) return fail(400, { error: 'A resolution summary is required.' });

		resolveIncident(params.id!, summary, form.get('alerts') === 'on', form.get('postmortem') === 'on');
	},

	reopen: async ({ params }) => {
		reopenIncident(params.id!);
	},

	link: async ({ request, params }) => {
		const form = await request.formData();
		linkIncident(
			params.id!,
			String(form.get('relation')) as RelationKind,
			String(form.get('incident'))
		);
	},

	entry: async ({ request, params }) => {
		const form = await request.formData();
		const retroTime = String(form.get('at') ?? '').trim();

		const at = retroTime
			? new Date(`${new Date().toISOString().slice(0, 10)}T${retroTime}:00Z`).toISOString()
			: undefined;

		addEntry(params.id!, String(form.get('type')) as EntryType, String(form.get('text') ?? ''), {
			at,
			retro: Boolean(retroTime)
		});
	},

	'add-follow-up': async ({ request, params }) => {
		const form = await request.formData();
		addFollowUp(
			params.id!,
			String(form.get('title') ?? ''),
			String(form.get('owner') ?? 'Maya Chen'),
			new Date(String(form.get('due'))).toISOString()
		);
	},

	'toggle-follow-up': async ({ request }) => {
		const form = await request.formData();
		toggleFollowUp(String(form.get('id')));
	},

	postmortem: async ({ request, params }) => {
		const form = await request.formData();
		advance(params.id!, String(form.get('state')) as 'draft' | 'in-review' | 'published');
	}
} satisfies Actions;
