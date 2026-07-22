import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import type {
	Alert,
	AlertEscalation,
	AlertSeverity,
	AlertStatus,
	EscalationEvent,
	EscalationEventKind,
	IngestionFailure,
	Silence,
	SilenceState
} from '$lib/alerts';
import { apiClient } from './api';

type Schemas = components['schemas'];

const HOUR = 60 * 60 * 1000;

function toLabels(labels: Record<string, string> | undefined): string[] {
	if (!labels) return [];
	return Object.entries(labels)
		.map(([key, value]) => `${key}:${value}`)
		.sort();
}

function isDeliveryEvent(kind: string): boolean {
	return kind === 'notified' || kind === 'push' || kind === 'sms' || kind === 'chat';
}

function deliveryTone(text: string, result: string): 'success' | 'warning' | 'brand' {
	if (text.endsWith('failed') || result === 'quiet hours' || result === 'no channel') return 'warning';
	return 'success';
}

function toTimeline(events: Schemas['AlertEvent'][] | undefined): EscalationEvent[] {
	return (events ?? []).map((event) => ({
		id: event.id,
		at: event.at,
		kind: event.kind as EscalationEventKind,
		text: event.text,
		result: event.result || undefined,
		tone:
			event.kind === 'resolved'
				? 'success'
				: event.kind === 'suppressed' || event.kind === 'timeout' || event.kind === 'exhausted'
					? 'warning'
					: event.kind === 'routed' || event.kind === 'escalation'
						? 'brand'
						: isDeliveryEvent(event.kind)
							? deliveryTone(event.text, event.result)
							: undefined
	}));
}

function toAlert(dto: Schemas['Alert']): Alert {
	return {
		id: dto.id,
		severity: dto.severity as AlertSeverity,
		title: dto.title,
		description: dto.description,
		source: dto.source,
		service: dto.service,
		status: dto.status as AlertStatus,
		ackedBy: dto.ackedBy ?? null,
		labels: toLabels(dto.labels),
		count: dto.count,
		firstSeenAt: dto.startedAt,
		lastSeenAt: dto.lastSeenAt,
		escalationPolicySlug: dto.escalationPolicySlug ?? null,
		escalation: dto.escalation
			? ({
					state: dto.escalation.state,
					stepIndex: dto.escalation.stepIndex,
					totalSteps: dto.escalation.totalSteps,
					cycle: dto.escalation.cycle,
					policySlug: dto.escalation.policySlug,
					nextAt: dto.escalation.nextAt ?? null,
					ackExpiresAt: dto.escalation.ackExpiresAt ?? null
				} satisfies AlertEscalation)
			: null,
		children: (dto.children ?? []).map((child) => ({
			id: child.id,
			title: child.title,
			count: child.count,
			lastSeenAt: child.lastSeenAt,
			status: child.status as AlertStatus,
			severity: child.severity as AlertSeverity
		})),
		links: (dto.links ?? []).map((link) => ({
			kind: link.kind as 'runbook' | 'dashboard' | 'source',
			label: link.label,
			url: link.url
		})),
		payload: dto.payload,
		timeline: toTimeline(dto.timeline)
	};
}

export type AlertQuery = {
	status?: string[];
	severity?: string[];
	source?: string[];
	service?: string[];
	label?: string[];
	since?: string;
	query?: string;
	cursor?: string;
	limit?: number;
};

export type AlertPage = {
	alerts: Alert[];
	nextCursor: string | null;
	facets: { sources: string[]; services: string[]; labels: string[] };
};

export async function listAlerts(
	cookies: Cookies,
	workspace: string,
	filter: AlertQuery = {}
): Promise<AlertPage> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/alerts', {
		params: { path: { workspaceId: workspace }, query: filter },
		querySerializer: { array: { style: 'form', explode: false } }
	});
	return {
		alerts: (data?.items ?? []).map(toAlert),
		nextCursor: data?.nextCursor || null,
		facets: data?.facets ?? { sources: [], services: [], labels: [] }
	};
}

export async function getAlert(
	cookies: Cookies,
	workspace: string,
	alertId: string
): Promise<Alert | undefined> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/alerts/{alertId}', {
		params: { path: { workspaceId: workspace, alertId } }
	});
	return data ? toAlert(data) : undefined;
}

export async function setStatus(
	cookies: Cookies,
	workspace: string,
	ids: string[],
	status: 'acked' | 'resolved'
): Promise<{ updated: number; error?: string }> {
	const { data, error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/alerts/status', {
		params: { path: { workspaceId: workspace } },
		body: { ids, status }
	});
	if (error) return { updated: 0, error: error.detail ?? 'Could not update those alerts.' };
	return { updated: data?.updated ?? 0 };
}

function toSilence(dto: Schemas['Silence']): Silence {
	return {
		id: dto.id,
		state: dto.state as SilenceState,
		scope: dto.conditions.map((c) => `${c.field}:${c.value}`),
		reason: dto.reason || 'No reason given',
		creator: dto.createdBy,
		startsAt: dto.startsAt,
		endsAt: dto.endsAt
	};
}

async function allSilences(cookies: Cookies, workspace: string): Promise<Silence[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/alert-silences', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map(toSilence);
}

export async function listSilences(cookies: Cookies, workspace: string): Promise<Silence[]> {
	return (await allSilences(cookies, workspace)).filter((silence) => silence.state !== 'ended');
}

export async function listSilenceHistory(cookies: Cookies, workspace: string): Promise<Silence[]> {
	return (await allSilences(cookies, workspace)).filter((silence) => silence.state === 'ended');
}

export async function createSilence(
	cookies: Cookies,
	workspace: string,
	input: {
		scope: string[];
		reason: string;
		startsNow: boolean;
		startsAt?: string;
		durationHours: number;
	}
): Promise<{ error?: string }> {
	const start = input.startsNow ? Date.now() : Date.parse(input.startsAt ?? '');
	if (Number.isNaN(start)) return { error: 'Give the silence a start time.' };

	const conditions = input.scope
		.map((entry) => {
			const [head, ...rest] = entry.split('=');
			const value = rest.join('=').trim();
			const parts = head.trim().split(/\s+/);
			const field = parts[0] as 'source' | 'service' | 'label';
			if (field === 'label') {
				const key = parts.slice(1).join(' ').trim();
				return { field, value: key ? `${key}=${value}` : value };
			}
			return { field, value };
		})
		.filter((condition) => ['source', 'service', 'label'].includes(condition.field) && condition.value);

	if (!conditions.length) return { error: 'A silence needs at least one scope.' };

	const { error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/alert-silences', {
		params: { path: { workspaceId: workspace } },
		body: {
			reason: input.reason,
			conditions,
			startsAt: new Date(start).toISOString(),
			endsAt: new Date(start + input.durationHours * HOUR).toISOString()
		}
	});
	return error ? { error: error.detail ?? 'Could not create the silence.' } : {};
}

export async function endSilence(
	cookies: Cookies,
	workspace: string,
	silenceId: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/alert-silences/{silenceId}/end',
		{ params: { path: { workspaceId: workspace, silenceId } } }
	);
	return error ? { error: error.detail ?? 'Could not end that silence.' } : {};
}

export async function listFailures(
	cookies: Cookies,
	workspace: string
): Promise<IngestionFailure[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/alert-failures', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map((failure) => ({
		id: failure.id,
		source: failure.source || 'unknown source',
		at: failure.at,
		reason: failure.detail || failure.reason,
		payload: failure.payload
	}));
}
