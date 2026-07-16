import { SEVERITY_TONE } from '$lib/dashboard';
import { patterns, postmortemId, waitingOn } from '$lib/postmortems';
import { listIncidents } from '$lib/server/incidents';
import { listPublished, publishedOn } from '$lib/server/postmortems';
import type { PageServerLoad } from './$types';

const DAY = 86_400_000;

export const load: PageServerLoad = ({ url }) => {
	const now = Date.now();

	const query = url.searchParams.get('q')?.toLowerCase() ?? '';
	const service = url.searchParams.get('service') ?? '';
	const severity = url.searchParams.get('severity') ?? '';
	const range = url.searchParams.get('range') ?? '90d';

	const published = listPublished();

	const within = (publishedAt: string | null) => {
		if (range === 'all' || !publishedAt) return true;
		const days = range === '30d' ? 30 : 90;
		return Date.parse(publishedAt) >= now - days * DAY;
	};

	const library = published
		.filter(({ postmortem, incident }) => {
			const haystack = `${incident.name} ${postmortem.id} ${incident.id}`.toLowerCase();
			return (
				(!query || haystack.includes(query)) &&
				(!service || incident.services.includes(service)) &&
				(!severity || incident.severity === severity) &&
				within(postmortem.publishedAt)
			);
		})
		.map((row) => ({
			id: row.postmortem.id,
			incidentId: row.incident.id,
			title: row.incident.name,
			severity: row.incident.severity,
			tone: SEVERITY_TONE[row.incident.severity],
			services: row.incident.services,
			author: row.postmortem.author,
			publishedAt: publishedOn(row)
		}));

	return {
		filters: { query: url.searchParams.get('q') ?? '', service, severity, range },
		anyPublished: published.length > 0,
		library,
		patterns: patterns(published.map((row) => row.postmortem)),
		waiting: listIncidents()
			.map((incident) => waitingOn(incident, now))
			.filter((row) => row !== null)
			.map((row) => ({
				...row,
				postmortemId: postmortemId(row.incidentId),
				tone: SEVERITY_TONE[row.severity]
			}))
	};
};
