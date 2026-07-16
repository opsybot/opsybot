import { error, fail } from '@sveltejs/kit';
import { PUBLISH_STAGES, STAGE_TONE, TEMPLATES, type ComponentState, type PublishStage } from '$lib/statuspages';
import { listPages, postStatusUpdate, publishIncident, publishTarget } from '$lib/server/statuspages';
import type { Actions, PageServerLoad } from './$types';

const STATES: ComponentState[] = ['operational', 'degraded', 'partial', 'major'];
const stageOf = (value: FormDataEntryValue | null): PublishStage | undefined =>
	PUBLISH_STAGES.find((stage) => stage === value);

export const load: PageServerLoad = ({ url }) => {
	const incident = publishTarget(url.searchParams.get('incident') ?? undefined);

	return {
		incident: incident
			? {
					id: incident.id,
					name: incident.name,
					title: incident.statusPage.title,
					services: incident.services,
					published: incident.statusPage.updates.map((update) => ({
						stage: update.stage.toLowerCase() as PublishStage,
						tone: STAGE_TONE[update.stage.toLowerCase() as PublishStage],
						text: update.text,
						at: update.at
					}))
				}
			: null,
		pages: listPages().map((page) => ({ id: page.id, name: page.name })),
		components: [
			...new Map(
				listPages()
					.flatMap((page) => page.components)
					.map((component) => [component.name, { name: component.name, services: component.services }])
			).values()
		],
		templates: TEMPLATES
	};
};

export const actions: Actions = {
	publish: async ({ request }) => {
		const form = await request.formData();

		const incidentId = String(form.get('incident'));
		const pageIds = form.getAll('page').map(String);
		const title = String(form.get('title') ?? '').trim();

		const componentStates: Record<string, ComponentState> = {};
		for (const name of form.getAll('component').map(String)) {
			const state = STATES.find((option) => option === form.get(`state:${name}`));
			componentStates[name] = state ?? 'degraded';
		}

		if (!pageIds.length) return fail(400, { error: 'Pick at least one page.' });
		if (!Object.keys(componentStates).length) return fail(400, { error: 'Pick at least one component.' });
		if (!title) return fail(400, { error: 'Write a public title customers will see.' });

		const ok = publishIncident({
			incidentId,
			pageIds,
			componentStates,
			title,
			stage: 'investigating',
			text: String(form.get('text') ?? '')
		});
		if (!ok) error(404, 'That incident no longer exists.');

		return { published: true, pages: pageIds.length };
	},

	update: async ({ request }) => {
		const form = await request.formData();
		const stage = stageOf(form.get('stage'));
		if (!stage) return fail(400, { error: 'Pick a stage.' });

		postStatusUpdate(String(form.get('incident')), stage, String(form.get('text') ?? ''));
		return { updated: true };
	}
};
