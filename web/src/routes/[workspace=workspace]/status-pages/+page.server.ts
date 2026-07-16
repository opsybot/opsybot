import { OVERALL, overallOf, subscriberTotal } from '$lib/statuspages';
import { listPages } from '$lib/server/statuspages';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	const pages = listPages();

	return {
		anyPage: pages.length > 0,
		pages: pages.map((page) => ({
			id: page.id,
			name: page.name,
			visibility: page.visibility,
			domain: page.domain,
			published: page.published,
			subscribers: subscriberTotal(page),
			overall: overallOf(page),
			overallLabel: OVERALL[overallOf(page)].label,
			overallTone: OVERALL[overallOf(page)].tone
		}))
	};
};
