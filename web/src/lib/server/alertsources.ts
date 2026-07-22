import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import { FORMATS, type Condition, type ConditionOp, type Mapping, type RoutingRule, type Source } from '$lib/alertsources';
import { formatSince, formatUtc } from '$lib/time';
import { apiClient } from './api';

type Schemas = components['schemas'];

import type { Tone } from '$lib/dashboard';

export type DeliveryEvent = { id: string; at: string; title: string; outcome: string; tone: Tone };

function iconFor(format: string): string {
	return FORMATS.find((entry) => entry.id === format)?.icon ?? 'braces';
}

function toSource(dto: Schemas['AlertSource']): Source {
	return {
		id: dto.slug,
		slug: dto.slug,
		name: dto.name,
		format: dto.format,
		icon: iconFor(dto.format),
		health: dto.health === 'paused' ? 'stale' : (dto.health as Source['health']),
		lastEvent: dto.lastEventAt ? formatSince(Date.now() - Date.parse(dto.lastEventAt)) : 'no events yet',
		status: dto.status as Source['status'],
		volume: [],
		failures: dto.failureCount,
		secret: dto.signingSecret ?? '',
		ingestUrl: dto.ingestUrl,
		mapping: (dto.mapping ?? []).map((m) => ({ field: m.field, path: m.path }))
	};
}

export function sanitizeMapping(raw: string): Mapping[] {
	let data: unknown;
	try {
		data = JSON.parse(raw);
	} catch {
		return [];
	}
	if (!Array.isArray(data)) return [];
	return data
		.filter(
			(row): row is Record<string, unknown> =>
				!!row && typeof row === 'object' && typeof row.field === 'string'
		)
		.map((row) => ({
			field: (row.field as string).trim(),
			path: typeof row.path === 'string' ? row.path : ''
		}))
		.filter((row) => row.field)
		.filter((row, index, all) => all.findIndex((other) => other.field === row.field) === index);
}

export async function listSources(cookies: Cookies, workspace: string): Promise<Source[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/alert-sources', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map(toSource);
}

export async function getSource(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<Source | undefined> {
	const { data } = await apiClient(cookies).GET(
		'/workspaces/{workspaceId}/alert-sources/{sourceSlug}',
		{ params: { path: { workspaceId: workspace, sourceSlug: slug } } }
	);
	return data ? toSource(data) : undefined;
}

export async function eventsFor(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<DeliveryEvent[]> {
	const { data } = await apiClient(cookies).GET(
		'/workspaces/{workspaceId}/alert-sources/{sourceSlug}/events',
		{ params: { path: { workspaceId: workspace, sourceSlug: slug }, query: { limit: 20 } } }
	);
	return (data?.items ?? []).map((event) => ({
		id: event.id,
		at: formatUtc(event.at),
		title: event.dedupKey || 'event',
		outcome: event.outcome,
		tone: (event.outcome === 'failed'
			? 'critical'
			: event.outcome === 'created'
				? 'brand'
				: 'neutral') as Tone
	}));
}

export async function createSource(
	cookies: Cookies,
	workspace: string,
	input: { name: string; formatId: string }
): Promise<{ slug?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/alert-sources', {
		params: { path: { workspaceId: workspace } },
		body: { name: input.name, format: input.formatId as Schemas['AlertSource']['format'] }
	});
	if (error) return { error: error.detail ?? 'Could not create the source.' };
	return { slug: data?.slug };
}

export async function setPaused(
	cookies: Cookies,
	workspace: string,
	slug: string,
	paused: boolean
): Promise<boolean> {
	const path = paused
		? '/workspaces/{workspaceId}/alert-sources/{sourceSlug}/pause'
		: '/workspaces/{workspaceId}/alert-sources/{sourceSlug}/resume';
	const { error } = await apiClient(cookies).POST(path, {
		params: { path: { workspaceId: workspace, sourceSlug: slug } }
	});
	return !error;
}

export async function rotateSecret(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<string | undefined> {
	const { data } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/alert-sources/{sourceSlug}/secret',
		{ params: { path: { workspaceId: workspace, sourceSlug: slug } } }
	);
	return data?.signingSecret;
}

export async function saveMapping(
	cookies: Cookies,
	workspace: string,
	slug: string,
	mapping: Mapping[]
): Promise<boolean> {
	const { error } = await apiClient(cookies).PUT(
		'/workspaces/{workspaceId}/alert-sources/{sourceSlug}/mapping',
		{
			params: { path: { workspaceId: workspace, sourceSlug: slug } },
			body: { mapping }
		}
	);
	return !error;
}

export async function deleteSource(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<boolean> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/alert-sources/{sourceSlug}',
		{ params: { path: { workspaceId: workspace, sourceSlug: slug } } }
	);
	return !error;
}

function toRule(dto: Schemas['AlertRoute']): RoutingRule {
	return {
		id: dto.id,
		policy: dto.policyRef,
		conditions: dto.conditions.map((c) => ({
			field: c.field,
			op: c.op as ConditionOp,
			value: c.value
		}))
	};
}

export async function listRules(
	cookies: Cookies,
	workspace: string
): Promise<{ rules: RoutingRule[]; defaultPolicy: string; knownPolicies: string[] }> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/alert-routes', {
		params: { path: { workspaceId: workspace } }
	});
	return {
		rules: (data?.items ?? []).map(toRule),
		defaultPolicy: data?.defaultPolicyRef ?? 'platform-default',
		knownPolicies: data?.knownPolicyRefs ?? []
	};
}

export async function setDefaultPolicy(
	cookies: Cookies,
	workspace: string,
	policy: string
): Promise<boolean> {
	const { error } = await apiClient(cookies).PUT('/workspaces/{workspaceId}/alert-routes/default', {
		params: { path: { workspaceId: workspace } },
		body: { policyRef: policy }
	});
	return !error;
}

export async function addRule(
	cookies: Cookies,
	workspace: string,
	rule: { conditions: Condition[]; policy: string }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/alert-routes', {
		params: { path: { workspaceId: workspace } },
		body: { policyRef: rule.policy, conditions: rule.conditions }
	});
	return error ? { error: error.detail ?? 'Could not save the rule.' } : {};
}

export async function updateRule(
	cookies: Cookies,
	workspace: string,
	id: string,
	rule: { conditions: Condition[]; policy: string }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).PUT(
		'/workspaces/{workspaceId}/alert-routes/{routeId}',
		{
			params: { path: { workspaceId: workspace, routeId: id } },
			body: { policyRef: rule.policy, conditions: rule.conditions }
		}
	);
	return error ? { error: error.detail ?? 'Could not save the rule.' } : {};
}

export async function deleteRule(
	cookies: Cookies,
	workspace: string,
	id: string
): Promise<boolean> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/alert-routes/{routeId}',
		{ params: { path: { workspaceId: workspace, routeId: id } } }
	);
	return !error;
}

export async function reorderRules(
	cookies: Cookies,
	workspace: string,
	ids: string[]
): Promise<boolean> {
	const { error } = await apiClient(cookies).PUT('/workspaces/{workspaceId}/alert-routes/order', {
		params: { path: { workspaceId: workspace } },
		body: { ids }
	});
	return !error;
}
