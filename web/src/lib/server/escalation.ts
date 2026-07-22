import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import {
	DEFAULT_HOURS,
	firstBranchKind,
	firstDeactivatedTarget,
	stepSummary,
	uid,
	type Branch,
	type Directory,
	type EscNode,
	type Hours,
	type Level,
	type SummaryPart,
	type Target,
	type TargetType,
	type Tree
} from '$lib/escalation';
import { apiClient } from './api';

type Schemas = components['schemas'];

export type PolicyListItem = {
	id: string;
	name: string;
	team: string;
	routed: number;
	branch: ReturnType<typeof firstBranchKind>;
	warning: string | null;
	summary: SummaryPart[];
};

export type RoutingLink = { id: string; rule: string };
export type RecentEscalation = {
	alertId: string;
	alert: string;
	at: string;
	outcome: string;
	tone: 'success' | 'critical' | 'neutral';
	by: string | null;
	duration: string;
};

function minutesToClock(minute: number): string {
	const h = String(Math.floor(minute / 60)).padStart(2, '0');
	const m = String(minute % 60).padStart(2, '0');
	return `${h}:${m}`;
}

function clockToMinutes(clock: string): number {
	const [h, m] = clock.split(':').map(Number);
	return (h || 0) * 60 + (m || 0);
}

function labelFor(directory: Directory, type: TargetType, ref: string): { label: string; invalid: boolean } {
	switch (type) {
		case 'person': {
			const m = directory.members.find((entry) => entry.id === ref);
			return { label: m?.name ?? 'removed member', invalid: !m || !m.active };
		}
		case 'schedule': {
			const s = directory.schedules.find((entry) => entry.id === ref);
			return { label: s?.slug ?? 'removed schedule', invalid: !s };
		}
		case 'team': {
			const t = directory.teams.find((entry) => entry.id === ref);
			return { label: t?.slug ?? 'removed team', invalid: !t };
		}
		case 'webhook': {
			const w = directory.webhooks.find((entry) => entry.id === ref);
			return { label: w?.slug ?? 'removed webhook', invalid: !w };
		}
	}
}

function apiHours(hours: Schemas['EscalationHours'] | undefined): Hours {
	if (!hours) return { ...DEFAULT_HOURS, days: [...DEFAULT_HOURS.days] };
	return {
		days: hours.days,
		start: minutesToClock(hours.startMinute),
		end: minutesToClock(hours.endMinute),
		timezone: hours.timezone
	};
}

function apiNodesToTree(nodes: Schemas['EscalationNode'][], directory: Directory): EscNode[] {
	return nodes.map((node): EscNode => {
		if (node.type === 'branch') {
			const branch: Branch = {
				id: node.id || uid('br'),
				type: 'branch',
				on: node.on === 'hours' ? 'hours' : 'priority',
				hours: apiHours(node.hours),
				lanes: (node.lanes ?? []).map((lane) => ({
					id: lane.id || uid('ln'),
					key: lane.key,
					nodes: apiNodesToTree(lane.nodes, directory)
				}))
			};
			return branch;
		}
		const level: Level = {
			id: node.id || uid('lv'),
			type: 'level',
			targets: (node.targets ?? []).map((t): Target => {
				const meta = labelFor(directory, t.type, t.ref);
				return { type: t.type, ref: t.ref, value: meta.label, invalid: meta.invalid };
			}),
			mode: node.mode === 'rr' ? 'rr' : 'all',
			wait: String(Math.max(1, Math.round((node.waitSeconds ?? 300) / 60))),
			addType: 'person'
		};
		return level;
	});
}

function treeNodesToApi(nodes: EscNode[]): Schemas['EscalationNode'][] {
	return nodes.map((node): Schemas['EscalationNode'] => {
		if (node.type === 'branch') {
			return {
				type: 'branch',
				id: node.id,
				on: node.on,
				hours:
					node.on === 'hours'
						? {
								days: node.hours.days,
								startMinute: clockToMinutes(node.hours.start),
								endMinute: clockToMinutes(node.hours.end),
								timezone: node.hours.timezone
							}
						: undefined,
				lanes: node.lanes.map((lane) => ({ id: lane.id, key: lane.key, nodes: treeNodesToApi(lane.nodes) }))
			};
		}
		return {
			type: 'level',
			id: node.id,
			targets: node.targets.map((t) => ({ type: t.type, ref: t.ref })),
			mode: node.mode,
			waitSeconds: (Number(node.wait) || 5) * 60
		};
	});
}

function toTree(policy: Schemas['EscalationPolicy'], directory: Directory): Tree {
	return {
		name: policy.name,
		team: policy.teamSlug,
		repeat: String(policy.repeat),
		ack: String(Math.round(policy.ackTimeoutSeconds / 60)),
		nodes: apiNodesToTree(policy.nodes, directory)
	};
}

function treeToRequest(tree: Tree): Schemas['SaveEscalationPolicyRequest'] {
	return {
		name: tree.name,
		teamSlug: tree.team || undefined,
		repeat: Number(tree.repeat) || 0,
		ackTimeoutSeconds: (Number(tree.ack) || 0) * 60,
		nodes: treeNodesToApi(tree.nodes)
	};
}

export async function getDirectory(cookies: Cookies, workspace: string): Promise<Directory> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/escalation-directory', {
		params: { path: { workspaceId: workspace } }
	});
	return data ?? { members: [], schedules: [], teams: [], webhooks: [] };
}

export async function listPolicies(
	cookies: Cookies,
	workspace: string,
	directory: Directory
): Promise<PolicyListItem[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/escalation-policies', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map((item) => {
		const tree: Tree = {
			name: item.name,
			team: item.teamSlug,
			repeat: '0',
			ack: '0',
			nodes: apiNodesToTree(item.nodes, directory)
		};
		const deactivated = firstDeactivatedTarget(tree);
		return {
			id: item.slug,
			name: item.name,
			team: item.teamSlug,
			routed: item.routed,
			branch: firstBranchKind(tree),
			warning: deactivated ? `references deactivated user ${deactivated.value}` : null,
			summary: stepSummary(tree)
		};
	});
}

export type PolicyDetail = {
	id: string;
	tree: Tree;
	routing: RoutingLink[];
	recent: RecentEscalation[];
	routed: number;
};

const RECENT_TONE: Record<string, 'success' | 'critical' | 'neutral'> = {
	acked: 'success',
	resolved: 'neutral',
	exhausted: 'critical',
	running: 'neutral'
};

function formatDuration(startIso: string, endIso: string | undefined): string {
	if (!endIso) return 'ongoing';
	const seconds = Math.max(0, Math.round((Date.parse(endIso) - Date.parse(startIso)) / 1000));
	const m = Math.floor(seconds / 60);
	const s = String(seconds % 60).padStart(2, '0');
	return `${m} m ${s} s`;
}

function ruleText(route: Schemas['AlertRoute']): string {
	return route.conditions.map((c) => `${c.field} ${c.op} ${c.value}`).join(' AND ');
}

export async function getPolicy(
	cookies: Cookies,
	workspace: string,
	slug: string,
	directory: Directory
): Promise<PolicyDetail | undefined> {
	const { data } = await apiClient(cookies).GET(
		'/workspaces/{workspaceId}/escalation-policies/{policySlug}',
		{ params: { path: { workspaceId: workspace, policySlug: slug } } }
	);
	if (!data) return undefined;
	const routing = data.routes.map((route) => ({ id: route.id, rule: ruleText(route) }));
	if (data.isDefault) {
		routing.push({ id: 'default-route', rule: 'Default route, catches everything unmatched' });
	}
	return {
		id: data.policy.slug,
		tree: toTree(data.policy, directory),
		routing,
		recent: data.recent.map((entry) => ({
			alertId: entry.alertId,
			alert: entry.alertTitle,
			at: entry.startedAt,
			outcome: entry.outcome,
			tone: RECENT_TONE[entry.state] ?? 'neutral',
			by: entry.by ?? null,
			duration: formatDuration(entry.startedAt, entry.endedAt)
		})),
		routed: data.routed
	};
}

export async function createPolicy(
	cookies: Cookies,
	workspace: string,
	tree: Tree
): Promise<{ slug?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/escalation-policies',
		{
			params: { path: { workspaceId: workspace } },
			body: treeToRequest(tree)
		}
	);
	if (error) return { error: error.detail ?? 'Could not save the policy.' };
	return { slug: data?.slug };
}

export async function updatePolicy(
	cookies: Cookies,
	workspace: string,
	slug: string,
	tree: Tree
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).PUT(
		'/workspaces/{workspaceId}/escalation-policies/{policySlug}',
		{
			params: { path: { workspaceId: workspace, policySlug: slug } },
			body: treeToRequest(tree)
		}
	);
	return error ? { error: error.detail ?? 'Could not save the policy.' } : {};
}

export async function deletePolicy(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/escalation-policies/{policySlug}',
		{ params: { path: { workspaceId: workspace, policySlug: slug } } }
	);
	return error ? { error: error.detail ?? 'Could not delete the policy.' } : {};
}

export type PolicyOption = { slug: string; name: string };

export async function listPolicyOptions(cookies: Cookies, workspace: string): Promise<PolicyOption[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/escalation-policies', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map((item) => ({ slug: item.slug, name: item.name }));
}

export type Webhook = { id: string; slug: string; name: string; url: string; hasSecret: boolean };

export async function listWebhooks(cookies: Cookies, workspace: string): Promise<Webhook[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/escalation-webhooks', {
		params: { path: { workspaceId: workspace } }
	});
	return data?.items ?? [];
}

export async function createWebhook(
	cookies: Cookies,
	workspace: string,
	input: { name: string; url: string; secret?: string }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/escalation-webhooks', {
		params: { path: { workspaceId: workspace } },
		body: { name: input.name, url: input.url, secret: input.secret || undefined }
	});
	return error ? { error: error.detail ?? 'Could not create the webhook.' } : {};
}

export async function deleteWebhook(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/escalation-webhooks/{webhookSlug}',
		{ params: { path: { workspaceId: workspace, webhookSlug: slug } } }
	);
	return error ? { error: error.detail ?? 'Could not delete the webhook.' } : {};
}

export async function escalateAlert(
	cookies: Cookies,
	workspace: string,
	alertId: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/alerts/{alertId}/escalate',
		{ params: { path: { workspaceId: workspace, alertId } } }
	);
	return error ? { error: error.detail ?? 'Could not escalate that alert.' } : {};
}
