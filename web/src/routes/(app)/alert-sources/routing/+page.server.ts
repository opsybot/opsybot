import { error, fail } from '@sveltejs/kit';
import { RT_FIELDS, RT_OPS, RT_POLICIES, RT_SAMPLE, type Condition, type ConditionOp } from '$lib/alertsources';
import {
	addRule,
	defaultPolicy,
	deleteRule,
	listRules,
	moveRule,
	setDefaultPolicy,
	updateRule
} from '$lib/server/alertsources';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = () => {
	return { rules: listRules(), defaultPolicy: defaultPolicy(), sample: RT_SAMPLE };
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
	saveRule: async ({ request }) => {
		const form = await request.formData();
		const parsed = parseRule(String(form.get('definition') ?? ''));
		if ('error' in parsed) return fail(400, { error: parsed.error });

		if (parsed.id) {
			if (!updateRule(parsed.id, { conditions: parsed.conditions, policy: parsed.policy })) {
				error(404, 'That rule no longer exists.');
			}
		} else {
			addRule({ conditions: parsed.conditions, policy: parsed.policy }, parsed.position);
		}
		return { saved: true };
	},

	deleteRule: async ({ request }) => {
		const form = await request.formData();
		deleteRule(String(form.get('id')));
		return { deleted: true };
	},

	moveRule: async ({ request }) => {
		const form = await request.formData();
		moveRule(String(form.get('id')), form.get('dir') === 'up' ? -1 : 1);
		return { moved: true };
	},

	setDefault: async ({ request }) => {
		const form = await request.formData();
		const policy = String(form.get('policy'));
		if (!RT_POLICIES.includes(policy)) return fail(400, { error: 'Unknown policy.' });
		setDefaultPolicy(policy);
		return { policy };
	}
};
