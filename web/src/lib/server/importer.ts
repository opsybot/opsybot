import type { CompareRow, CutoverSource, DryRun, ImportPlan } from '$lib/importer';

const DRYRUN: DryRun = {
	created: [
		{ kind: 'schedules', n: 6, note: 'rotations, layers, and overrides' },
		{ kind: 'escalation policies', n: 4, note: 'steps and targets mapped 1:1' },
		{ kind: 'teams', n: 5, note: 'with membership' },
		{ kind: 'users', n: 42, note: 'matched by email' },
		{ kind: 'services', n: 11, note: 'from Opsgenie services' }
	],
	decisions: [
		{
			id: 'd1',
			kind: 'user',
			title: '3 users have no matching email in Opsybot',
			detail: "They're referenced by schedules. Invite them, or map to an existing member.",
			choices: [
				{ value: 'invite', label: 'Invite the 3 users' },
				{ value: 'map', label: 'Map to existing members' }
			]
		},
		{
			id: 'd2',
			kind: 'integration',
			title: 'Integration "Zabbix webhook" has no Opsybot equivalent',
			detail: 'Opsybot ingests Zabbix via generic JSON. Map it, or skip and reconnect manually after.',
			choices: [
				{ value: 'map', label: 'Map to generic JSON' },
				{ value: 'skip', label: 'Skip: reconnect later' }
			]
		},
		{
			id: 'd3',
			kind: 'routing',
			title: 'Routing rule uses a tag filter Opsybot expresses differently',
			detail: 'Opsgenie "responders" tag → Opsybot label matcher. Confirm the translation.',
			choices: [
				{ value: 'translate', label: 'Use the translation' },
				{ value: 'skip', label: 'Skip this rule' }
			]
		}
	],
	skipped: [
		{
			title: 'Opsgenie "Who is on-call" widget config',
			reason: "UI-only, no Opsybot equivalent: schedules carry over, the widget doesn't."
		},
		{ title: '2 deactivated users', reason: 'Not referenced anywhere; nothing to import.' }
	]
};

const COMPARE: CompareRow[] = [
	{ schedule: 'payments-primary', opsy: 'Priya Nair', og: 'Priya Nair', match: true },
	{ schedule: 'platform-default', opsy: 'Dev Patel', og: 'Dev Patel', match: true },
	{ schedule: 'frontend-daytime', opsy: 'Sana Ito', og: 'Sana Ito', match: true },
	{ schedule: 'security-oncall', opsy: 'Maya Chen', og: 'Maya Chen', match: true },
	{ schedule: 'db-escalation', opsy: 'Marcus Lee', og: 'Marcus Lee', match: true }
];

const CUTOVER: CutoverSource[] = [
	{ source: 'Prometheus Alertmanager', from: 'api.opsgenie.com/v2/…', to: 'in.opsy.bot/e/acme/am-prod' },
	{ source: 'Grafana', from: 'opsgenie contact point', to: 'in.opsy.bot/e/acme/grafana' },
	{ source: 'Datadog', from: 'Opsgenie integration', to: 'in.opsy.bot/e/acme/datadog' },
	{ source: 'AWS CloudWatch (SNS)', from: 'Opsgenie SNS endpoint', to: 'in.opsy.bot/e/acme/cloudwatch' }
];

export function getImportPlan(): ImportPlan {
	return { dryrun: DRYRUN, compare: COMPARE, cutover: CUTOVER };
}
