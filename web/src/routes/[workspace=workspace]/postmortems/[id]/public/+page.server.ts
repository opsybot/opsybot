import { error } from '@sveltejs/kit';
import { formatDuration, impactWindow } from '$lib/postmortems';
import { listFollowUps } from '$lib/server/incidents';
import { getPostmortem } from '$lib/server/postmortems';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const found = getPostmortem(params.id);
	if (!found) error(404, `No postmortem called ${params.id}.`);

	const { postmortem, incident } = found;

	return {
		id: postmortem.id,
		title: incident.name,
		organization: params.workspace.toLowerCase(),
		date: (postmortem.publishedAt ?? incident.declaredAt).slice(0, 10),
		window: impactWindow(incident),
		resolved: incident.status === 'resolved',
		published: incident.postmortem === 'published',
		live: postmortem.publicLink,
		facts: [
			{ label: 'Duration', value: formatDuration(incident) },
			...incident.customFields.map((field) => ({ label: field.label, value: field.value }))
		],
		summary: postmortem.summary,
		impact: postmortem.impact,
		factors: postmortem.factors.map((factor) => factor.text).filter(Boolean),
		changes: listFollowUps()
			.filter((followUp) => followUp.incidentId === incident.id)
			.map((followUp) => followUp.title)
	};
};
