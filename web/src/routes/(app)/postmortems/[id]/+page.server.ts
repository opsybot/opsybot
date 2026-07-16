import { error, fail } from '@sveltejs/kit';
import { PEOPLE } from '$lib/incidents';
import { facts, type SectionId } from '$lib/postmortems';
import { addFollowUp, listFollowUps } from '$lib/server/incidents';
import {
	addFactor,
	draft,
	getPostmortem,
	openPostmortem,
	publish,
	removeFactor,
	requestReview,
	setOptions,
	writeFactor,
	writeSection
} from '$lib/server/postmortems';
import type { Actions, PageServerLoad } from './$types';

const SECTIONS: SectionId[] = ['summary', 'impact', 'wentWell', 'improve'];

const sectionOf = (value: FormDataEntryValue | null): SectionId | undefined =>
	SECTIONS.find((section) => section === value);

function open(pmId: string) {
	const found = getPostmortem(pmId) ?? openPostmortem(pmId.replace('PM-', 'INC-'));
	if (!found) error(404, `No postmortem called ${pmId}.`);
	return found;
}

function editable(pmId: string) {
	const found = open(pmId);
	if (found.incident.postmortem === 'published') {
		error(403, 'This postmortem is published and can no longer be edited.');
	}
	return found;
}

export const load: PageServerLoad = ({ params }) => {
	const { postmortem, incident } = open(params.id);

	return {
		postmortem,
		incidentId: incident.id,
		title: incident.name,
		state: incident.postmortem,
		facts: facts(incident),
		followUps: listFollowUps().filter((followUp) => followUp.incidentId === incident.id),
		people: PEOPLE
	};
};

export const actions: Actions = {
	draft: async ({ request, params }) => {
		const form = await request.formData();
		const section = sectionOf(form.get('section'));
		if (!section) return fail(400, { error: 'Unknown section.' });

		const { incident } = editable(params.id);
		return { text: draft(incident, section) };
	},

	section: async ({ request, params }) => {
		const form = await request.formData();
		const section = sectionOf(form.get('section'));
		if (!section) return fail(400, { error: 'Unknown section.' });

		editable(params.id);
		writeSection(params.id, section, String(form.get('text') ?? ''));
	},

	addFactor: async ({ params }) => {
		editable(params.id);
		addFactor(params.id);
	},

	factor: async ({ request, params }) => {
		const form = await request.formData();
		editable(params.id);

		// One field per submit; label and text autosave independently
		const field = form.get('field') === 'label' ? 'label' : 'text';
		const value = String(form.get('value') ?? '');
		writeFactor(params.id, String(form.get('id')), {
			[field]: field === 'label' ? value.trim() : value
		});
	},

	removeFactor: async ({ request, params }) => {
		const form = await request.formData();
		editable(params.id);
		removeFactor(params.id, String(form.get('id')));
	},

	options: async ({ request, params }) => {
		const form = await request.formData();
		const option = String(form.get('option'));
		const on = form.get('on') === 'true';

		if (option !== 'announce' && option !== 'publicLink') return fail(400, { error: 'Unknown option.' });

		// Options stay editable after publishing, so open() rather than editable()
		open(params.id);
		setOptions(params.id, { [option]: on });
	},

	followUp: async ({ request, params }) => {
		const form = await request.formData();
		const title = String(form.get('title') ?? '').trim();
		if (!title) return fail(400, { error: 'Give the follow-up a title.' });

		const { incident } = editable(params.id);
		addFollowUp(incident.id, title, String(form.get('owner')), String(form.get('due')));

		return { added: true };
	},

	review: async ({ request, params }) => {
		const form = await request.formData();
		const reviewer = String(form.get('reviewer'));

		const { incident } = open(params.id);
		if (incident.postmortem === 'published' || incident.postmortem === 'in-review') {
			error(409, 'This postmortem is already past review.');
		}

		requestReview(params.id, reviewer);
		return { reviewer };
	},

	publish: async ({ params }) => {
		const { postmortem, incident } = open(params.id);
		if (incident.postmortem === 'published') error(409, 'This postmortem is already published.');

		publish(params.id);
		return { published: true, announced: postmortem.announce };
	}
};
