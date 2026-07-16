import { formatGap, gaps, handovers } from '$lib/oncall';
import { listSchedules, onCallNow } from '$lib/server/oncall';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	const now = new Date();

	return {
		now: now.getTime(),
		schedules: listSchedules().map((schedule) => {
			const [gap] = gaps(schedule, now, 7);
			const [next] = handovers(schedule, now, 14, 1);

			return {
				id: schedule.id,
				name: schedule.name,
				team: schedule.team,
				paused: schedule.paused,
				gap: gap ? formatGap(gap) : null,
				handover: next,
				...onCallNow(schedule, now)
			};
		})
	};
};
