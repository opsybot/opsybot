import { error, fail, redirect } from '@sveltejs/kit';
import {
	addOverride,
	archiveSchedule,
	deleteSchedule,
	duplicateSchedule,
	resumeSchedule,
	scheduleDetail,
	unarchiveSchedule
} from '$lib/server/oncall';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url, cookies }) => {
	const view = url.searchParams.get('view') === 'month' ? 'month' : 'week';
	const zone = url.searchParams.get('tz') === 'local' ? 'local' : 'utc';
	const date = url.searchParams.get('date') ?? undefined;
	const time = url.searchParams.get('time') ?? undefined;

	const detail = await scheduleDetail(cookies, params.workspace, params.id, { view, zone, date, time });
	if (!detail) error(404, `No schedule called ${params.id}.`);
	return detail;
};

export const actions: Actions = {
	override: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const full = form.get('mode') !== 'partial';

		const startsAt = full
			? String(form.get('targetStart'))
			: `${form.get('startDate')}T${form.get('startTime')}:00Z`;
		const endsAt = full
			? String(form.get('targetEnd'))
			: `${form.get('endDate')}T${form.get('endTime')}:00Z`;

		if (Number.isNaN(Date.parse(startsAt)) || Number.isNaN(Date.parse(endsAt))) {
			return fail(400, { error: 'Give the override a start and an end.' });
		}

		const outcome = await addOverride(cookies, params.workspace, params.id, {
			person: String(form.get('person')),
			startsAt,
			endsAt,
			reason: String(form.get('reason') ?? '').trim() || 'No reason given'
		});

		if (outcome.error) return fail(400, { error: outcome.error });
	},

	duplicate: async ({ params, cookies }) => {
		const copy = await duplicateSchedule(cookies, params.workspace, params.id);
		if (!copy.slug) error(400, copy.error ?? `Could not duplicate ${params.id}.`);
		redirect(303, `/${params.workspace}/on-call/${copy.slug}`);
	},

	resume: async ({ params, cookies }) => {
		await resumeSchedule(cookies, params.workspace, params.id);
	},

	archive: async ({ params, cookies }) => {
		await archiveSchedule(cookies, params.workspace, params.id);
		redirect(303, `/${params.workspace}/on-call`);
	},

	unarchive: async ({ params, cookies }) => {
		if (!(await unarchiveSchedule(cookies, params.workspace, params.id)))
			return fail(400, { error: 'Could not restore the schedule.' });
	},

	delete: async ({ params, cookies }) => {
		const outcome = await deleteSchedule(cookies, params.workspace, params.id);
		if (outcome.error) return fail(400, { error: outcome.error });
		redirect(303, `/${params.workspace}/on-call`);
	}
};
