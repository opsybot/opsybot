import { error, fail } from '@sveltejs/kit';
import { RT_FIELDS, RT_OPS, RT_POLICIES, RT_SAMPLE, type Condition, type ConditionOp } from '$lib/alertsources';
import {
	addRule,
	deleteRule,
	listRules,
	reorderRules,
	setDefaultPolicy,
	updateRule
} from '$lib/server/alertsources';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const { rules, defaultPolicy, knownPolicies } = await listRules(cookies, params.workspace);
	return { rules, defaultPolicy, knownPolicies, sample: RT_SAMPLE };
};

function parseRule(raw: string): { id: string | null; conditions: Condition[]; policy: string; position: string } | { error: string } {
	let data: unknown;
	try {
		data = JSON.parse(raw);
	} catch {
		return { error: 'Could not read the rule.' };
	}
	if (!data || typeof data !== 'object') return { error: 'Could not read the rule.' };
	const object = data as Record<string, unknown>;

	const conditions: Condition[] = (Array.isArray(object.conditions) ? object.conditions : [])
		.filter((entry): entry is Record<string, unknown> => !!entry && typeof entry === 'object')
		.filter(
			(entry) =>
				typeof entry.field === 'string' &&
				RT_FIELDS.includes(entry.field) &&
				typeof entry.op === 'string' &&
				RT_OPS.includes(entry.op as ConditionOp) &&
				typeof entry.value === 'string' &&
				entry.value.trim()
		)
		.map((entry) => ({
			field: entry.field as string,
			op: entry.op as ConditionOp,
			value: (entry.value as string).trim()
		}));

	if (!conditions.length) return { error: 'A rule needs at least one condition with a value.' };

	const policy = typeof object.policy === 'string' && RT_POLICIES.includes(object.policy) ? object.policy : RT_POLICIES[0];
	const position = typeof object.position === 'string' ? object.position : 'end';
	const id = typeof object.id === 'string' ? object.id : null;
	return { id, conditions, policy, position };
}

export const actions: Actions = {
	saveRule: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const parsed = parseRule(String(form.get('definition') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });

		const rule = { conditions: parsed.conditions, policy: parsed.policy };
		const outcome = parsed.id
			? await updateRule(cookies, params.workspace, parsed.id, rule)
			: await addRule(cookies, params.workspace, rule);
		if (outcome.error) return fail(400, { error: outcome.error });
		return { saved: true };
	},

	deleteRule: async ({ request, params, cookies }) => {
		const form = await request.formData();
		if (!(await deleteRule(cookies, params.workspace, String(form.get('id'))))) {
			return fail(400, { error: 'Could not delete that rule.' });
		}
		return { deleted: true };
	},

	moveRule: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const id = String(form.get('id'));
		const { rules } = await listRules(cookies, params.workspace);
		const ids = rules.map((rule) => rule.id);
		const index = ids.indexOf(id);
		const target = form.get('dir') === 'up' ? index - 1 : index + 1;
		if (index === -1 || target < 0 || target >= ids.length) return { moved: false };

		[ids[index], ids[target]] = [ids[target], ids[index]];
		if (!(await reorderRules(cookies, params.workspace, ids))) {
			return fail(400, { error: 'Could not reorder the rules.' });
		}
		return { moved: true };
	},

	setDefault: async ({ request, params, cookies }) => {
		const form = await request.formData();
		const policy = String(form.get('policy'));
		if (!(await setDefaultPolicy(cookies, params.workspace, policy))) {
			return fail(400, { error: 'Could not set the default policy.' });
		}
		return { policy };
	}
};
