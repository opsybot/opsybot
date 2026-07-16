export type TriggerType = 'declared' | 'severity' | 'status' | 'resolved' | 'overdue';

export const TRIGGERS: { value: TriggerType; label: string }[] = [
	{ value: 'declared', label: 'Incident declared' },
	{ value: 'severity', label: 'Severity changed' },
	{ value: 'status', label: 'Status changed' },
	{ value: 'resolved', label: 'Incident resolved' },
	{ value: 'overdue', label: 'Update overdue' }
];

export const TRIGGER_LABEL: Record<TriggerType, string> = {
	declared: 'Incident declared',
	severity: 'Severity changed',
	status: 'Status changed',
	resolved: 'Incident resolved',
	overdue: 'Update overdue'
};

export function isTriggerType(value: string): value is TriggerType {
	return TRIGGERS.some((trigger) => trigger.value === value);
}

export type ActionType = 'post' | 'role' | 'note' | 'webhook' | 'followup' | 'statuspage';

export const ACTION_TYPES: { value: ActionType; label: string; icon: string }[] = [
	{ value: 'post', label: 'Post message', icon: 'megaphone' },
	{ value: 'role', label: 'Prompt role assignment', icon: 'users' },
	{ value: 'note', label: 'Add timeline note', icon: 'file-text' },
	{ value: 'webhook', label: 'Call webhook', icon: 'webhook' },
	{ value: 'followup', label: 'Create follow-up', icon: 'list-checks' },
	{ value: 'statuspage', label: 'Prompt status page update', icon: 'globe' }
];

export const ACTION_META: Record<ActionType, { label: string; icon: string }> = Object.fromEntries(
	ACTION_TYPES.map((action) => [action.value, { label: action.label, icon: action.icon }])
) as Record<ActionType, { label: string; icon: string }>;

export function isActionType(value: string): value is ActionType {
	return ACTION_TYPES.some((action) => action.value === value);
}

export type Condition = { field: string; value: string };

export type WorkflowAction = { id: string; type: ActionType; config: Record<string, string> };

export type WorkflowRun = {
	id: string;
	at: string;
	incident: string;
	summary: string;
	ok: boolean;
	error?: string;
	retriable?: boolean;
};

export type Workflow = {
	id: string;
	name: string;
	enabled: boolean;
	trigger: TriggerType;
	conditions: Condition[];
	actions: WorkflowAction[];
	// Newest first; lastRun reads the head
	history: WorkflowRun[];
};

export type Role = {
	id: string;
	name: string;
	description: string;
	builtin?: boolean;
};

export type Template = {
	id: string;
	name: string;
	icon: string;
	description: string;
	trigger: TriggerType;
	conditions: Condition[];
	actions: { type: ActionType; config: Record<string, string> }[];
};

export const TEMPLATES: Template[] = [
	{
		id: 'tpl1',
		name: 'SEV1 comms cadence',
		icon: 'megaphone',
		description: 'On SEV1 declare: announce, assign comms, start the 15-min update clock.',
		trigger: 'declared',
		conditions: [{ field: 'severity', value: 'SEV1' }],
		actions: [
			{
				type: 'post',
				config: {
					channel: '#incidents',
					text: 'SEV1 declared: {name}. Channel: {channel}. Lead: {lead}. Next update in 15 min.'
				}
			},
			{ type: 'role', config: { role: 'Comms lead' } },
			{ type: 'note', config: { text: 'SEV1 comms cadence active — updates every 15 min.' } }
		]
	},
	{
		id: 'tpl2',
		name: 'Security auto-invite',
		icon: 'shield-check',
		description: 'Security-labeled incidents pull the security team in automatically.',
		trigger: 'declared',
		conditions: [{ field: 'label security', value: 'true' }],
		actions: [
			{
				type: 'post',
				config: { channel: '#security', text: 'Security incident {id} declared — you have been added to {channel}.' }
			},
			{ type: 'webhook', config: { url: 'https://hooks.acme.dev/sec-pager', payload: '{ "incident": "{id}" }' } }
		]
	},
	{
		id: 'tpl3',
		name: 'Postmortem follow-up auto-create',
		icon: 'list-checks',
		description: 'On resolve of SEV1/SEV2: create the postmortem follow-up with a due date.',
		trigger: 'resolved',
		conditions: [{ field: 'severity', value: 'SEV1, SEV2' }],
		actions: [
			{ type: 'followup', config: { title: 'Write postmortem for {id}', owner: 'incident lead', due: 'in 3 working days' } },
			{ type: 'note', config: { text: 'Postmortem follow-up created automatically.' } }
		]
	}
];

export function getTemplate(id: string): Template | undefined {
	return TEMPLATES.find((template) => template.id === id);
}

export const CONDITION_FIELDS = ['severity', 'service', 'label security', 'label region'];

export const POST_CHANNELS = ['#incidents', '#eng-all', '#ops', '#security', 'the incident channel'];
export const FOLLOWUP_OWNERS = ['incident lead', 'comms lead', 'Maya Chen', 'Priya Nair'];
export const FOLLOWUP_DUES = ['in 1 working day', 'in 3 working days', 'in 1 week'];
export const STATUS_PAGE_OPTIONS = ['status.acme.dev', 'internal.acme.dev'];

export function shortId(prefix: string): string {
	return prefix + Math.random().toString(36).slice(2, 8);
}

export function defaultConfig(type: ActionType): Record<string, string> {
	switch (type) {
		case 'post':
			return { channel: '#incidents', text: '' };
		case 'role':
			return { role: 'Comms lead' };
		case 'note':
			return { text: '' };
		case 'webhook':
			return { url: '', payload: '{ "incident": "{id}" }' };
		case 'followup':
			return { title: '', owner: 'incident lead', due: 'in 3 working days' };
		case 'statuspage':
			return { page: 'status.acme.dev' };
	}
}

export function describeTrigger(trigger: TriggerType, conditions: Condition[]): string {
	const parts = [TRIGGER_LABEL[trigger].toLowerCase()];
	for (const condition of conditions) {
		if (condition.field.trim() && condition.value.trim()) {
			parts.push(`${condition.field.trim()} is ${condition.value.trim()}`);
		}
	}
	return parts.join(' · ');
}

export function lastRun(workflow: Pick<Workflow, 'history'>): WorkflowRun | null {
	return workflow.history[0] ?? null;
}

// A timeline note re-arms the overdue timer, so this pairing fires forever
export function loops(trigger: TriggerType, actions: Pick<WorkflowAction, 'type'>[]): boolean {
	return trigger === 'overdue' && actions.some((action) => action.type === 'note');
}

export function webhookMissingUrl(action: WorkflowAction): boolean {
	return action.type === 'webhook' && !action.config.url?.trim();
}

export type WorkflowInput = {
	name: string;
	trigger: TriggerType;
	conditions: Condition[];
	actions: WorkflowAction[];
};

export function workflowErrors(input: WorkflowInput): string[] {
	const errors: string[] = [];
	if (!input.name.trim()) errors.push('Give the workflow a name.');
	if (input.actions.length === 0) errors.push('A workflow needs at least one action.');
	if (loops(input.trigger, input.actions)) {
		errors.push('This workflow would loop: an “update overdue” trigger with a timeline-note action re-fires itself.');
	}
	if (input.actions.some(webhookMissingUrl)) errors.push('Every webhook action needs a URL.');
	return errors;
}
