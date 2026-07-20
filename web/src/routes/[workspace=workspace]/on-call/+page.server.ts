import { scheduleList } from '$lib/server/oncall';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ cookies, params }) => {
	return scheduleList(cookies, params.workspace);
};
