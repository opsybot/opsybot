import { fail } from '@sveltejs/kit';
import { createMonitor, deleteMonitor, listMonitors, updateMonitor } from '$lib/server/monitors';
import { listRules } from '$lib/server/alertsources';
import { listPolicyOptions } from '$lib/server/escalation';
import type { Actions, PageServerLoad } from './$types';

function readForm(form: FormData): {
	name: string;
	intervalSeconds: number;
	graceSeconds: number;
	policyRef: string;
} {
	return {
		name: String(form.get('name') ?? '').trim(),
		intervalSeconds: Number(form.get('interval') ?? 0),
		graceSeconds: Number(form.get('grace') ?? 0),
		policyRef: String(form.get('policy') ?? '').trim()
	};
}

export const load: PageServerLoad = async ({ params, cookies }) => {
	const [heartbeats, routing, policies] = await Promise.all([
		listMonitors(cookies, params.workspace),
		listRules(cookies, params.workspace),
		listPolicyOptions(cookies, params.workspace)
	]);
	return {
		now: Date.now(),
		heartbeats,
		knownPolicies: policies.map((p) => p.slug),
		defaultPolicy: routing.defaultPolicy
	};
};

export const actions: Actions = {
	create: async ({ request, params, cookies }) => {
		const input = readForm(await request.formData());
		if (!input.name) return fail(400, { error: 'Give the monitor a name.' });

		const { monitor, error } = await createMonitor(cookies, params.workspace, input);
		if (error || !monitor) return fail(400, { error: error ?? 'Could not create that monitor.' });
		return { url: monitor.checkInUrl };
	},

	update: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const input = readForm(form);
		if (!input.name) return fail(400, { error: 'Give the monitor a name.' });

		const { error } = await updateMonitor(cookies, params.workspace, String(form.get('id')), input);
		if (error) return fail(400, { error });
		return { saved: true };
	},

	delete: async ({ request, params, cookies }) => {
		const form = await request.formData();
		if (!(await deleteMonitor(cookies, params.workspace, String(form.get('id'))))) {
			return fail(400, { error: 'Could not delete that monitor.' });
		}
		return { deleted: true };
	}
};
