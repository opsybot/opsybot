import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import type { Severity, Tone } from '$lib/dashboard';
import { SEVERITY_TONE } from '$lib/dashboard';
import type {
	EntryType,
	FollowUp,
	Incident,
	IncidentStage,
	LinkedAlert,
	RelationKind,
	TimelineEntry,
	TimelineRevision
} from '$lib/incidents';
import { apiClient, apiFetch } from './api';

type Schemas = components['schemas'];
type IncidentDto = Schemas['Incident'];

const ALERT_SEVERITY: Record<string, { code: string; tone: Tone }> = {
	critical: { code: 'CRIT', tone: 'critical' },
	high: { code: 'HIGH', tone: 'high' },
	warning: { code: 'WARN', tone: 'warning' }
};

const RELATION_TO_FRONT: Record<string, RelationKind> = {
	related: 'related to',
	duplicate: 'duplicates',
	caused_by: 'caused by'
};

const RELATION_TO_API: Record<RelationKind, 'related' | 'duplicate' | 'caused_by'> = {
	'related to': 'related',
	duplicates: 'duplicate',
	'caused by': 'caused_by'
};

const FIELD_KIND_TO_FRONT: Record<string, string> = {
	text: 'text',
	select: 'select',
	multi_select: 'multi-select',
	number: 'number'
};

const FIELD_KIND_TO_API: Record<string, 'text' | 'select' | 'multi_select' | 'number'> = {
	text: 'text',
	select: 'select',
	'multi-select': 'multi_select',
	number: 'number'
};

function severityTone(level: string): Tone {
	return (SEVERITY_TONE as Record<string, Tone>)[level] ?? 'neutral';
}

function alertSeverity(sev: string): { code: string; tone: Tone } {
	return ALERT_SEVERITY[sev] ?? { code: sev.toUpperCase(), tone: 'neutral' };
}

function toTimelineEntry(e: Schemas['IncidentEvent']): TimelineEntry {
	return {
		id: e.id,
		type: e.category as EntryType,
		at: e.at,
		actor: e.alertId ? (e.alertTitle ?? 'Linked alert') : (e.actor ?? 'Opsybot'),
		text: e.text,
		editable: e.kind === 'note' && !e.alertId,
		attachments: (e.attachments ?? []).map((a) => ({
			id: a.id,
			entryId: a.entryId,
			kind: a.kind,
			label: a.label,
			url: a.url,
			body: a.body,
			sizeBytes: a.sizeBytes
		})),
		ai: !e.actor && !e.alertId,
		retro: e.retroactive,
		edited: !!e.editedAt,
		editedAt: e.editedAt,
		alertId: e.alertId,
		alertTitle: e.alertTitle,
		result: e.result
	};
}

function toCustomFields(
	map: Record<string, string> | undefined,
	defs: Schemas['IncidentField'][]
): { label: string; value: string }[] {
	if (!map) return [];
	const byId = new Map(defs.filter((d) => d.id).map((d) => [d.id as string, d]));
	return Object.entries(map).flatMap(([defId, value]) => {
		const def = byId.get(defId);
		return def ? [{ label: def.name, value }] : [];
	});
}

function toIncident(
	dto: IncidentDto,
	defs: Schemas['IncidentField'][],
	currentUserId: string
): Incident {
	return {
		id: dto.id,
		ref: `INC-${dto.number}`,
		name: dto.name,
		severity: dto.severityLevel as Severity,
		status: dto.status as IncidentStage,
		lead: dto.leadLabel ?? '',
		leadUserId: dto.leadUserId ?? '',
		comms: '',
		team: dto.teamSlug ?? '',
		services: (dto.services ?? []).map((s) => s.name),
		declaredAt: dto.declaredAt,
		nextUpdateAt: null,
		resolvedAt: dto.resolvedAt ?? null,
		mine:
			!!currentUserId && (dto.leadUserId === currentUserId || dto.declaredBy === currentUserId),
		summary: dto.summary ?? '',
		customFields: toCustomFields(dto.customFields, defs),
		customFieldsRaw: dto.customFields ?? {},
		related: (dto.relations ?? []).map((r) => ({
			relation: RELATION_TO_FRONT[r.kind] ?? 'related to',
			id: `INC-${r.relatedNumber}`,
			name: r.relatedName,
			relationId: r.id
		})),
		alerts: (dto.alerts ?? []).map((a) => {
			const s = alertSeverity(a.severity);
			return {
				id: a.alertId,
				title: a.title,
				severity: s.code,
				tone: s.tone,
				status: a.status as LinkedAlert['status']
			};
		}),
		timeline: (dto.timeline ?? []).map(toTimelineEntry),
		statusPage: { domain: '', stage: 'none', title: '', updates: [] },
		postmortem: 'not-started'
	};
}

function toFollowUp(dto: Schemas['IncidentFollowup']): FollowUp {
	return {
		id: dto.id,
		incidentId: dto.incidentId,
		title: dto.title,
		owner: dto.ownerLabel ?? '',
		dueAt: dto.dueAt ?? '',
		done: dto.done
	};
}

export type IncidentQuery = {
	status?: string[];
	severity?: string[];
	service?: string[];
	team?: string[];
	active?: boolean;
	since?: string;
	query?: string;
	cursor?: string;
	limit?: number;
};

export async function meId(cookies: Cookies): Promise<string> {
	const { data } = await apiClient(cookies).GET('/me');
	return data?.id ?? '';
}

export async function listMembers(
	cookies: Cookies,
	workspace: string
): Promise<{ id: string; name: string }[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/members', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? [])
		.filter((m) => m.status === 'active')
		.map((m) => ({ id: m.userId, name: m.name }));
}

export async function listIncidents(
	cookies: Cookies,
	workspace: string,
	filter: IncidentQuery = {},
	currentUserId = ''
): Promise<{ incidents: Incident[]; nextCursor: string | null }> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/incidents', {
		params: { path: { workspaceId: workspace }, query: filter },
		querySerializer: { array: { style: 'form', explode: false } }
	});
	return {
		incidents: (data?.items ?? []).map((dto) => toIncident(dto, [], currentUserId)),
		nextCursor: data?.nextCursor || null
	};
}

async function fetchIncidentDto(
	cookies: Cookies,
	workspace: string,
	id: string
): Promise<IncidentDto | undefined> {
	const { data } = await apiClient(cookies).GET(
		'/workspaces/{workspaceId}/incidents/{incidentId}',
		{ params: { path: { workspaceId: workspace, incidentId: id } } }
	);
	return data;
}

export async function getIncident(
	cookies: Cookies,
	workspace: string,
	id: string
): Promise<Incident | undefined> {
	const [dto, fields] = await Promise.all([
		fetchIncidentDto(cookies, workspace, id),
		fieldDefs(cookies, workspace)
	]);
	return dto ? toIncident(dto, fields, '') : undefined;
}

export async function getIncidentDetail(
	cookies: Cookies,
	workspace: string,
	id: string
): Promise<{ incident: Incident; followUps: FollowUp[] } | undefined> {
	const [dto, fields] = await Promise.all([
		fetchIncidentDto(cookies, workspace, id),
		fieldDefs(cookies, workspace)
	]);
	if (!dto) return undefined;
	return {
		incident: toIncident(dto, fields, ''),
		followUps: (dto.followups ?? []).map(toFollowUp)
	};
}

async function fieldDefs(cookies: Cookies, workspace: string): Promise<Schemas['IncidentField'][]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/incident-fields', {
		params: { path: { workspaceId: workspace } }
	});
	return data?.items ?? [];
}

export async function listOpenFollowUps(cookies: Cookies, workspace: string): Promise<FollowUp[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/incident-followups', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map(toFollowUp);
}

export async function declareIncident(
	cookies: Cookies,
	workspace: string,
	input: {
		name: string;
		severityLevel: string;
		serviceIds: string[];
		leadUserId: string;
		alertIds: string[];
	}
): Promise<{ id?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST('/workspaces/{workspaceId}/incidents', {
		params: { path: { workspaceId: workspace } },
		body: {
			name: input.name,
			severityLevel: input.severityLevel || undefined,
			serviceIds: input.serviceIds,
			leadUserId: input.leadUserId || undefined
		}
	});
	if (error || !data) return { error: error?.detail ?? 'Could not declare the incident.' };
	for (const alertId of input.alertIds) {
		await apiClient(cookies).POST(
			'/workspaces/{workspaceId}/incidents/{incidentId}/alerts',
			{ params: { path: { workspaceId: workspace, incidentId: data.id } }, body: { alertId } }
		);
	}
	return { id: data.id };
}

export async function declareFromAlert(
	cookies: Cookies,
	workspace: string,
	input: { alertId: string; name?: string; severityLevel?: string }
): Promise<{ id?: string; error?: string }> {
	const { data, error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/from-alert',
		{
			params: { path: { workspaceId: workspace } },
			body: {
				alertId: input.alertId,
				name: input.name || undefined,
				severityLevel: input.severityLevel || undefined
			}
		}
	);
	if (error || !data) return { error: error?.detail ?? 'Could not declare the incident.' };
	return { id: data.id };
}

export async function updateIncident(
	cookies: Cookies,
	workspace: string,
	id: string,
	patch: { name?: string; summary?: string; leadUserId?: string; teamSlug?: string; serviceIds?: string[] }
): Promise<{ error?: string }> {
	const current = await fetchIncidentDto(cookies, workspace, id);
	if (!current) return { error: 'That incident no longer exists.' };
	const { error } = await apiClient(cookies).PATCH(
		'/workspaces/{workspaceId}/incidents/{incidentId}',
		{
			params: { path: { workspaceId: workspace, incidentId: id } },
			body: {
				name: patch.name ?? current.name,
				summary: patch.summary ?? current.summary ?? '',
				teamSlug: patch.teamSlug ?? current.teamSlug ?? '',
				leadUserId: patch.leadUserId ?? current.leadUserId ?? '',
				serviceIds: patch.serviceIds ?? (current.services ?? []).map((s) => s.id)
			}
		}
	);
	return error ? { error: error.detail ?? 'Could not update the incident.' } : {};
}

export async function changeSeverity(
	cookies: Cookies,
	workspace: string,
	id: string,
	level: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/severity',
		{ params: { path: { workspaceId: workspace, incidentId: id } }, body: { level } }
	);
	return error ? { error: error.detail ?? 'Could not change the severity.' } : {};
}

export async function moveStatus(
	cookies: Cookies,
	workspace: string,
	id: string,
	status: IncidentStage
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/status',
		{ params: { path: { workspaceId: workspace, incidentId: id } }, body: { status } }
	);
	return error ? { error: error.detail ?? "That status change isn't allowed." } : {};
}

export async function resolveIncident(
	cookies: Cookies,
	workspace: string,
	id: string,
	summary: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/resolve',
		{ params: { path: { workspaceId: workspace, incidentId: id } }, body: { summary } }
	);
	return error ? { error: error.detail ?? 'Could not resolve the incident.' } : {};
}

export async function reopenIncident(
	cookies: Cookies,
	workspace: string,
	id: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/reopen',
		{ params: { path: { workspaceId: workspace, incidentId: id } } }
	);
	return error ? { error: error.detail ?? 'Could not reopen the incident.' } : {};
}

export async function setCustomFields(
	cookies: Cookies,
	workspace: string,
	id: string,
	fields: Record<string, string>
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).PUT(
		'/workspaces/{workspaceId}/incidents/{incidentId}/custom-fields',
		{ params: { path: { workspaceId: workspace, incidentId: id } }, body: { fields } }
	);
	return error ? { error: error.detail ?? 'Could not update custom fields.' } : {};
}

export async function linkAlert(
	cookies: Cookies,
	workspace: string,
	id: string,
	alertId: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/alerts',
		{ params: { path: { workspaceId: workspace, incidentId: id } }, body: { alertId } }
	);
	return error ? { error: error.detail ?? 'Could not link the alert.' } : {};
}

export async function unlinkAlert(
	cookies: Cookies,
	workspace: string,
	id: string,
	alertId: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/incidents/{incidentId}/alerts/{alertId}',
		{ params: { path: { workspaceId: workspace, incidentId: id, alertId } } }
	);
	return error ? { error: error.detail ?? 'Could not unlink the alert.' } : {};
}

export async function unrelateIncident(
	cookies: Cookies,
	workspace: string,
	id: string,
	relationId: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/incidents/{incidentId}/relations/{relationId}',
		{ params: { path: { workspaceId: workspace, incidentId: id, relationId } } }
	);
	return error ? { error: error.detail ?? 'Could not remove the relation.' } : {};
}

export async function relateIncident(
	cookies: Cookies,
	workspace: string,
	id: string,
	relation: RelationKind,
	relatedId: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/relations',
		{
			params: { path: { workspaceId: workspace, incidentId: id } },
			body: { relatedId, kind: RELATION_TO_API[relation] ?? 'related' }
		}
	);
	return error ? { error: error.detail ?? 'Could not link the incident.' } : {};
}

export async function addFollowUp(
	cookies: Cookies,
	workspace: string,
	id: string,
	input: { title: string; ownerUserId: string; dueAt?: string }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/followups',
		{
			params: { path: { workspaceId: workspace, incidentId: id } },
			body: {
				title: input.title,
				ownerUserId: input.ownerUserId || undefined,
				dueAt: input.dueAt || undefined
			}
		}
	);
	return error ? { error: error.detail ?? 'Could not add the follow-up.' } : {};
}

export async function toggleFollowUp(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	followupId: string,
	done: boolean
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/followups/{followupId}',
		{
			params: { path: { workspaceId: workspace, incidentId, followupId } },
			body: { done }
		}
	);
	return error ? { error: error.detail ?? 'Could not update the follow-up.' } : {};
}

export async function listTimeline(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	options: { category?: EntryType[]; cursor?: string; limit?: number } = {}
): Promise<{ entries: TimelineEntry[]; nextCursor: string }> {
	const { data } = await apiClient(cookies).GET(
		'/workspaces/{workspaceId}/incidents/{incidentId}/timeline',
		{
			params: {
				path: { workspaceId: workspace, incidentId },
				query: {
					category: options.category?.length ? options.category : undefined,
					cursor: options.cursor || undefined,
					limit: options.limit
				}
			}
		}
	);
	return {
		entries: (data?.entries ?? []).map(toTimelineEntry),
		nextCursor: data?.nextCursor ?? ''
	};
}

export async function addTimelineEntry(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	entry: { text: string; category: EntryType; at?: string; idempotencyKey?: string }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/timeline',
		{
			params: { path: { workspaceId: workspace, incidentId } },
			body: {
				text: entry.text,
				category: entry.category,
				at: entry.at,
				idempotencyKey: entry.idempotencyKey
			}
		}
	);
	return error ? { error: error.detail ?? 'Could not add the entry.' } : {};
}

export async function editTimelineEntry(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	entryId: string,
	entry: { text: string; category: EntryType }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).PATCH(
		'/workspaces/{workspaceId}/incidents/{incidentId}/timeline/{entryId}',
		{
			params: { path: { workspaceId: workspace, incidentId, entryId } },
			body: { text: entry.text, category: entry.category }
		}
	);
	return error ? { error: error.detail ?? 'Could not save the entry.' } : {};
}

export async function listEntryRevisions(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	entryId: string
): Promise<TimelineRevision[]> {
	const { data } = await apiClient(cookies).GET(
		'/workspaces/{workspaceId}/incidents/{incidentId}/timeline/{entryId}/revisions',
		{ params: { path: { workspaceId: workspace, incidentId, entryId } } }
	);
	return (data?.revisions ?? []).map((r) => ({
		id: r.id,
		at: r.at,
		editor: r.editorLabel ?? '',
		text: r.text,
		type: r.category as EntryType
	}));
}

export async function addAttachment(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	entryId: string,
	attachment: { kind: 'log' | 'link'; label: string; url?: string; body?: string }
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).POST(
		'/workspaces/{workspaceId}/incidents/{incidentId}/timeline/{entryId}/attachments',
		{
			params: { path: { workspaceId: workspace, incidentId, entryId } },
			body: attachment
		}
	);
	return error ? { error: error.detail ?? 'Could not attach the evidence.' } : {};
}

export async function uploadImageAttachment(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	entryId: string,
	label: string,
	file: File
): Promise<{ error?: string }> {
	const form = new FormData();
	form.set('label', label);
	form.set('file', file);
	const response = await apiFetch(
		cookies,
		`/workspaces/${encodeURIComponent(workspace)}/incidents/${encodeURIComponent(incidentId)}/timeline/${encodeURIComponent(entryId)}/attachments`,
		{ method: 'POST', body: form }
	);
	if (response.ok) return {};
	const problem = await response.json().catch(() => null);
	return { error: problem?.detail ?? 'Could not upload the image.' };
}

export function openAttachmentContent(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	attachmentId: string
) {
	return apiFetch(
		cookies,
		`/workspaces/${encodeURIComponent(workspace)}/incidents/${encodeURIComponent(incidentId)}/attachments/${encodeURIComponent(attachmentId)}/content`
	);
}

export async function removeAttachment(
	cookies: Cookies,
	workspace: string,
	incidentId: string,
	attachmentId: string
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).DELETE(
		'/workspaces/{workspaceId}/incidents/{incidentId}/attachments/{attachmentId}',
		{ params: { path: { workspaceId: workspace, incidentId, attachmentId } } }
	);
	return error ? { error: error.detail ?? 'Could not remove the attachment.' } : {};
}

export async function exportTimeline(
	cookies: Cookies,
	workspace: string,
	incidentId: string
): Promise<{ json: string; text: string } | { error: string }> {
	const { data, error } = await apiClient(cookies).GET(
		'/workspaces/{workspaceId}/incidents/{incidentId}/timeline/export',
		{ params: { path: { workspaceId: workspace, incidentId } } }
	);
	if (error || !data) return { error: error?.detail ?? 'Could not export the timeline.' };
	return { json: JSON.stringify(data, null, 2), text: data.text };
}

export type SeverityConfig = { id: string; def: string; label: string; tone: Tone; position: number };

export async function listSeverities(cookies: Cookies, workspace: string): Promise<SeverityConfig[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/incident-severities', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map((s, index) => ({
		id: s.level,
		def: s.definition ?? '',
		label: s.label,
		tone: (s.tone as Tone) ?? severityTone(s.level),
		position: s.position ?? index
	}));
}

export async function saveSeverityDefs(
	cookies: Cookies,
	workspace: string,
	defs: string[]
): Promise<{ error?: string }> {
	const current = await listSeverities(cookies, workspace);
	const severities = current.map((sev, index) => ({
		level: sev.id,
		label: sev.label,
		definition: (defs[index] ?? sev.def).slice(0, 200),
		tone: sev.tone,
		position: index
	}));
	const { error } = await apiClient(cookies).PUT('/workspaces/{workspaceId}/incident-severities', {
		params: { path: { workspaceId: workspace } },
		body: { severities }
	});
	return error ? { error: error.detail ?? 'Could not save the severities.' } : {};
}

export type FieldConfig = { id: string; name: string; type: string; options?: string };

export async function listFields(cookies: Cookies, workspace: string): Promise<FieldConfig[]> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/incident-fields', {
		params: { path: { workspaceId: workspace } }
	});
	return (data?.items ?? []).map((f) => ({
		id: f.id ?? '',
		name: f.name,
		type: FIELD_KIND_TO_FRONT[f.kind] ?? 'text',
		options: (f.options ?? []).join(', ') || undefined
	}));
}

function saveFieldsBody(fields: FieldConfig[]) {
	return {
		fields: fields.map((f) => ({
			id: f.id || undefined,
			name: f.name,
			kind: FIELD_KIND_TO_API[f.type] ?? 'text',
			options: f.options
				? f.options
						.split(',')
						.map((o) => o.trim())
						.filter(Boolean)
				: []
		}))
	};
}

export async function addField(
	cookies: Cookies,
	workspace: string,
	field: { name: string; type: string; options?: string }
): Promise<{ error?: string }> {
	const current = await listFields(cookies, workspace);
	const next = [...current, { id: '', name: field.name, type: field.type, options: field.options }];
	const { error } = await apiClient(cookies).PUT('/workspaces/{workspaceId}/incident-fields', {
		params: { path: { workspaceId: workspace } },
		body: saveFieldsBody(next)
	});
	return error ? { error: error.detail ?? 'Could not add the field.' } : {};
}

export async function removeField(
	cookies: Cookies,
	workspace: string,
	id: string
): Promise<{ error?: string }> {
	const current = await listFields(cookies, workspace);
	const next = current.filter((f) => f.id !== id);
	if (next.length === current.length) return { error: 'That field no longer exists.' };
	const { error } = await apiClient(cookies).PUT('/workspaces/{workspaceId}/incident-fields', {
		params: { path: { workspaceId: workspace } },
		body: saveFieldsBody(next)
	});
	return error ? { error: error.detail ?? 'Could not remove the field.' } : {};
}
