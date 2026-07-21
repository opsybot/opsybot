import type { ActionType, Condition, Role, TriggerType, Workflow, WorkflowAction } from '$lib/workflows';
import { CONDITION_FIELDS, isActionType, isTriggerType, shortId, workflowErrors } from '$lib/workflows';
import { scenario } from './fixtures';

function seed() {
	const roles: Role[] = [
		{
			id: 'lead',
			name: 'Incident lead',
			builtin: true,
			description:
				'Owns the incident end to end: sets direction, makes the calls, keeps the timeline honest. Every incident has exactly one.'
		},
		{
			id: 'comms',
			name: 'Comms lead',
			description:
				'Owns everything the outside world sees: status page updates, stakeholder pings, the next-update clock.'
		},
		{
			id: 'scribe',
			name: 'Scribe',
			description:
				'Keeps the timeline complete while others fix things. Opsybot drafts entries; the scribe verifies and fills gaps.'
		},
		{
			id: 'liaison',
			name: 'Customer liaison',
			description:
				'Talks to support and key accounts. Translates incident-speak into what customers need to hear.'
		}
	];

	const workflows: Workflow[] = [
		{
			id: 'wf1',
			name: 'SEV1 comms cadence',
			enabled: true,
			trigger: 'declared',
			conditions: [{ field: 'severity', value: 'SEV1' }],
			actions: [
				{
					id: 'wf1-a1',
					type: 'post',
					config: {
						channel: '#incidents',
						text: 'SEV1 declared: {name}. Channel: {channel}. Lead: {lead}. Next update in 15 min.'
					}
				},
				{ id: 'wf1-a2', type: 'role', config: { role: 'Comms lead' } },
				{ id: 'wf1-a3', type: 'note', config: { text: 'SEV1 comms cadence active: updates every 15 min.' } }
			],
			history: [
				{
					id: 'wf1-r1',
					at: '2026-07-11T09:14:00Z',
					incident: 'INC-2481',
					summary: 'Posted kickoff to #incidents · prompted comms-lead assignment · set 15-min update reminder',
					ok: true
				},
				{
					id: 'wf1-r2',
					at: '2026-07-04T22:40:00Z',
					incident: 'INC-2477',
					summary: 'Posted kickoff to #incidents · prompted comms-lead assignment · set 15-min update reminder',
					ok: true
				}
			]
		},
		{
			id: 'wf2',
			name: 'Security auto-invite',
			enabled: true,
			trigger: 'declared',
			conditions: [{ field: 'label security', value: 'true' }],
			actions: [
				{
					id: 'wf2-a1',
					type: 'post',
					config: {
						channel: '#security',
						text: 'Security incident {id} declared. You have been added to {channel}.'
					}
				},
				{
					id: 'wf2-a2',
					type: 'webhook',
					config: { url: 'https://hooks.acme.dev/sec-pager', payload: '{ "incident": "{id}" }' }
				}
			],
			history: [
				{
					id: 'wf2-r1',
					at: '2026-06-28T03:12:00Z',
					incident: 'INC-2465',
					summary: 'Added security team to channel · webhook sec-pager',
					ok: false,
					error: 'webhook sec-pager returned 503',
					retriable: true
				},
				{
					id: 'wf2-r2',
					at: '2026-06-14T11:05:00Z',
					incident: 'INC-2452',
					summary: 'Added security team to channel · webhook sec-pager',
					ok: true
				}
			]
		},
		{
			id: 'wf3',
			name: 'Postmortem follow-up nudge',
			enabled: false,
			trigger: 'resolved',
			conditions: [{ field: 'severity', value: 'SEV1 or SEV2' }],
			actions: [
				{
					id: 'wf3-a1',
					type: 'followup',
					config: { title: 'Write postmortem for {id}', owner: 'incident lead', due: 'in 3 working days' }
				},
				{ id: 'wf3-a2', type: 'note', config: { text: 'Postmortem follow-up created automatically.' } }
			],
			history: []
		}
	];

	return { roles, workflows };
}

const store = seed();
if (scenario() === 'empty') store.workflows = [];

export function listWorkflows(): Workflow[] {
	return store.workflows;
}

export function getWorkflow(id: string): Workflow | undefined {
	return store.workflows.find((workflow) => workflow.id === id);
}

export type WorkflowDefinition = {
	name: string;
	trigger: TriggerType;
	conditions: Condition[];
	actions: WorkflowAction[];
};

export function createWorkflow(definition: WorkflowDefinition): Workflow {
	const workflow: Workflow = {
		id: shortId('wf'),
		name: definition.name.trim(),
		enabled: false,
		trigger: definition.trigger,
		conditions: definition.conditions,
		actions: definition.actions,
		history: []
	};
	store.workflows.push(workflow);
	return workflow;
}

export function updateWorkflow(id: string, definition: WorkflowDefinition): boolean {
	const workflow = getWorkflow(id);
	if (!workflow) return false;
	workflow.name = definition.name.trim();
	workflow.trigger = definition.trigger;
	workflow.conditions = definition.conditions;
	workflow.actions = definition.actions;
	return true;
}

export function setEnabled(id: string, enabled: boolean): boolean {
	const workflow = getWorkflow(id);
	if (!workflow) return false;
	workflow.enabled = enabled;
	return true;
}

export function parseDefinition(raw: string): { definition: WorkflowDefinition } | { error: string } {
	let data: unknown;
	try {
		data = JSON.parse(raw);
	} catch {
		return { error: 'Could not read the workflow.' };
	}
	if (!data || typeof data !== 'object') return { error: 'Could not read the workflow.' };

	const object = data as Record<string, unknown>;
	const trigger = typeof object.trigger === 'string' && isTriggerType(object.trigger) ? object.trigger : null;
	if (!trigger) return { error: 'Pick a trigger for the workflow.' };

	const conditions: Condition[] = (Array.isArray(object.conditions) ? object.conditions : [])
		.filter((entry): entry is Record<string, unknown> => !!entry && typeof entry === 'object')
		.map((entry) => ({
			field: typeof entry.field === 'string' ? entry.field.trim() : '',
			value: typeof entry.value === 'string' ? entry.value.trim() : ''
		}))
		.filter((condition) => CONDITION_FIELDS.includes(condition.field) && condition.value);

	const actions: WorkflowAction[] = (Array.isArray(object.actions) ? object.actions : [])
		.filter((entry): entry is Record<string, unknown> => !!entry && typeof entry === 'object')
		.filter((entry) => typeof entry.type === 'string' && isActionType(entry.type))
		.map((entry) => ({
			id: typeof entry.id === 'string' && entry.id ? entry.id : shortId('a'),
			type: entry.type as ActionType,
			config: stringConfig(entry.config)
		}));

	const definition: WorkflowDefinition = {
		name: typeof object.name === 'string' ? object.name : '',
		trigger,
		conditions,
		actions
	};

	const errors = workflowErrors(definition);
	if (errors.length) return { error: errors[0] };
	return { definition };
}

function stringConfig(value: unknown): Record<string, string> {
	if (!value || typeof value !== 'object') return {};
	const out: Record<string, string> = {};
	for (const [key, raw] of Object.entries(value as Record<string, unknown>)) {
		if (typeof raw === 'string') out[key] = raw;
	}
	return out;
}

export function retryRun(workflowId: string, runId: string): boolean {
	const workflow = getWorkflow(workflowId);
	const run = workflow?.history.find((entry) => entry.id === runId);
	if (!run || !run.retriable) return false;
	run.ok = true;
	run.retriable = false;
	delete run.error;
	return true;
}

export function listRoles(): Role[] {
	return store.roles;
}

export function addRole(input: { name: string; description: string }): Role {
	const role: Role = { id: shortId('r'), name: input.name.trim(), description: input.description.trim() };
	store.roles.push(role);
	return role;
}

export function updateRole(id: string, input: { name: string; description: string }): boolean {
	const role = store.roles.find((entry) => entry.id === id);
	if (!role) return false;
	role.name = input.name.trim();
	role.description = input.description.trim();
	return true;
}

export function removeRole(id: string): boolean {
	const role = store.roles.find((entry) => entry.id === id);
	if (!role || role.builtin) return false;
	store.roles = store.roles.filter((entry) => entry.id !== id);
	return true;
}
