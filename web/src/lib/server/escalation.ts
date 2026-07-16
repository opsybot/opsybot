import type { Level, Tree } from '$lib/escalation';
import { firstBranchKind, firstDeactivatedTarget, stepSummary } from '$lib/escalation';
import { scenario } from './fixtures';

export type RoutingLink = { rule: string; matched: string };
export type RecentEscalation = {
	id: string | null;
	alert: string;
	at: string;
	outcome: string;
	tone: 'success' | 'critical' | 'neutral';
	by: string | null;
	duration: string;
};

export type Policy = {
	id: string;
	routed: number;
	tree: Tree;
	routing: RoutingLink[];
	recent: RecentEscalation[];
};

const level = (over: Partial<Level> & { id: string }): Level => ({
	type: 'level',
	targets: [],
	mode: 'all',
	wait: '5',
	addType: 'person',
	...over
});

function seed(): Policy[] {
	return [
		{
			id: 'payments-primary',
			routed: 14,
			tree: {
				name: 'payments-primary',
				team: 'payments',
				repeat: '2',
				nodes: [
					level({ id: 'lv-oncall', wait: '5', targets: [{ type: 'schedule', value: 'payments-primary' }] }),
					{
						id: 'br-pri',
						type: 'branch',
						on: 'priority',
						lanes: [
							{
								id: 'ln-hi',
								key: 'high',
								nodes: [
									level({ id: 'lv-priya', wait: '5', targets: [{ type: 'person', value: 'Priya Nair' }] }),
									level({ id: 'lv-team', wait: '10', mode: 'all', targets: [{ type: 'team', value: 'payments' }] })
								]
							},
							{
								id: 'ln-lo',
								key: 'low',
								nodes: [
									level({
										id: 'lv-leads',
										wait: '15',
										mode: 'rr',
										targets: [
											{ type: 'person', value: 'Marcus Lee' },
											{ type: 'person', value: 'Dev Patel' }
										]
									})
								]
							}
						]
					}
				]
			},
			routing: [
				{ rule: 'source = Checkly AND label team = payments', matched: '27 alerts / 30 d' },
				{ rule: 'service = payments-api', matched: '61 alerts / 30 d' },
				{ rule: 'label severity = critical AND label region = eu-*', matched: '9 alerts / 30 d' }
			],
			recent: [
				{ id: 'INC-2481', alert: 'Synthetic checkout failing from us-east-1', at: '2026-07-11 09:12 UTC', outcome: 'acked at level 2', tone: 'success', by: 'Priya Nair', duration: '7 m 08 s' },
				{ id: null, alert: 'payments-api p99 above 800 ms', at: '2026-07-11 06:40 UTC', outcome: 'acked at level 1', tone: 'success', by: 'Maya Chen', duration: '1 m 44 s' },
				{ id: null, alert: 'Queue depth above 10k', at: '2026-07-10 22:15 UTC', outcome: 'exhausted — unacked', tone: 'critical', by: null, duration: '35 m' },
				{ id: null, alert: 'Disk usage 85% on db-3', at: '2026-07-10 14:02 UTC', outcome: 'resolved before level 2', tone: 'neutral', by: null, duration: '4 m 51 s' }
			]
		},
		{
			id: 'platform-default',
			routed: 22,
			tree: {
				name: 'platform-default',
				team: 'platform',
				repeat: '1',
				nodes: [
					level({ id: 'pd-sched', wait: '5', targets: [{ type: 'schedule', value: 'platform-default' }] }),
					level({
						id: 'pd-leads',
						wait: '15',
						mode: 'rr',
						targets: [
							{ type: 'person', value: 'Marcus Lee' },
							{ type: 'person', value: 'Dev Patel' }
						]
					}),
					level({ id: 'pd-hook', wait: '5', targets: [{ type: 'webhook', value: 'ops-bridge' }] })
				]
			},
			routing: [
				{ rule: 'label team = platform', matched: '41 alerts / 30 d' },
				{ rule: 'service = gateway', matched: '18 alerts / 30 d' }
			],
			recent: [
				{ id: null, alert: 'Gateway 5xx rate above 2%', at: '2026-07-11 03:22 UTC', outcome: 'acked at level 1', tone: 'success', by: 'Marcus Lee', duration: '2 m 30 s' },
				{ id: null, alert: 'Build pipeline stalled', at: '2026-07-09 19:48 UTC', outcome: 'acked at level 2', tone: 'success', by: 'Dev Patel', duration: '18 m 04 s' }
			]
		},
		{
			id: 'frontend-daytime',
			routed: 6,
			tree: {
				name: 'frontend-daytime',
				team: 'frontend',
				repeat: '0',
				nodes: [
					level({ id: 'fd-tom', wait: '10', targets: [{ type: 'person', value: 'Tom Weber (deactivated)' }] }),
					level({ id: 'fd-sched', wait: '5', targets: [{ type: 'schedule', value: 'frontend-daytime' }] })
				]
			},
			routing: [{ rule: 'label team = frontend', matched: '6 alerts / 30 d' }],
			recent: [
				{ id: null, alert: 'LCP regression on /checkout', at: '2026-07-08 11:05 UTC', outcome: 'resolved before level 2', tone: 'neutral', by: null, duration: '9 m 12 s' }
			]
		}
	];
}

const store = { policies: seed() };
if (scenario() === 'empty') store.policies = [];

// Route segments beside [id]; a policy id must not shadow them
const RESERVED = new Set(['new']);

const slugify = (name: string): string =>
	name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');

export function listPolicies() {
	return store.policies.map((policy) => {
		const deactivated = firstDeactivatedTarget(policy.tree);
		return {
			id: policy.id,
			name: policy.tree.name,
			team: policy.tree.team,
			routed: policy.routed,
			branch: firstBranchKind(policy.tree),
			warning: deactivated
				? `references deactivated user ${deactivated.value.replace(/\s*\(deactivated\)/i, '')}`
				: null,
			summary: stepSummary(policy.tree)
		};
	});
}

export function getPolicy(id: string): Policy | undefined {
	return store.policies.find((policy) => policy.id === id);
}

export function updateTree(id: string, tree: Tree): boolean {
	const policy = getPolicy(id);
	if (!policy) return false;
	policy.tree = tree;
	return true;
}

export function createPolicy(tree: Tree): Policy {
	const base = slugify(tree.name) || 'policy';
	let unique = base;
	for (let n = 2; RESERVED.has(unique) || store.policies.some((policy) => policy.id === unique); n++) {
		unique = `${base}-${n}`;
	}
	const policy: Policy = { id: unique, routed: 0, tree, routing: [], recent: [] };
	store.policies.push(policy);
	return policy;
}
