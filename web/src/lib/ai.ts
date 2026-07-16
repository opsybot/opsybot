export type Health = 'ok' | 'fail';

export type Model = {
	id: string;
	name: string;
	endpoint: string;
	health: Health;
	latency: string;
};

export type AiFeatureId = 'summaries' | 'postmortems' | 'correlation';

export type AiFeature = { id: AiFeatureId; label: string; desc: string };

export type PromptSpec = { fields: string[]; excluded: string[]; template: string };

export const AI_FEATURES: AiFeature[] = [
	{ id: 'summaries', label: 'Incident summaries', desc: '"Catch me up" drawers and the pinned overview summary.' },
	{ id: 'postmortems', label: 'Postmortem drafts', desc: 'Section drafts assembled from the timeline.' },
	{ id: 'correlation', label: 'Alert correlation', desc: 'Grouping suggestions and related-incident hints.' }
];

// {{name}} placeholders are literal; template literals only interpolate ${...}
export const AI_PROMPTS: Record<AiFeatureId, PromptSpec> = {
	summaries: {
		fields: [
			'incident title + severity + status',
			'timeline entries (type, UTC time, actor, text)',
			'linked alert titles',
			'affected service names'
		],
		excluded: ['user emails', 'notification channels', 'API keys and endpoint secrets', 'subscriber data'],
		template: `You are the incident scribe. Summarize the incident below for a responder
who just joined. Be factual and terse. Use only information present in the
timeline; never speculate about cause. Times in UTC.

Incident: {{title}} ({{severity}}, {{status}})
Services: {{services}}
Linked alerts: {{alert_titles}}

Timeline:
{{timeline_entries}}

Write: 3-5 sentences. What happened, current impact, what has been done,
what is in progress. No names unless a decision is attributed.`
	},
	postmortems: {
		fields: [
			'incident record (severity, duration, impact window)',
			'timeline entries',
			'resolution summary',
			'follow-up items'
		],
		excluded: ['user emails', 'chat transcripts outside the incident channel', 'credentials'],
		template: `Draft the "{{section}}" section of a blameless postmortem.
Describe systems and conditions, never individuals' mistakes.
Use only facts from the timeline below. Mark uncertain inferences
with "likely". Times in UTC.

{{incident_facts}}
{{timeline_entries}}`
	},
	correlation: {
		fields: [
			'open alert titles + labels + service',
			'first/last seen timestamps',
			'service dependency edges (from catalog)'
		],
		excluded: ['alert payloads', 'user data', 'runbook contents'],
		template: `Given the open alerts and the service dependency graph, propose
groups of alerts that likely share a cause. For each group give:
shared service or dependency path, temporal proximity, confidence
(low/medium/high). Do not propose fixes.

Alerts:
{{open_alerts}}

Dependencies:
{{dependency_edges}}`
	}
};

export const TIMEOUT_OPTIONS = ['10 s', '30 s', '60 s'];
export const CONTEXT_OPTIONS = ['32k tokens', '128k tokens', '200k tokens'];

export const USE_DEFAULT = 'default';

const FEATURE_IDS = new Set<string>(AI_FEATURES.map((feature) => feature.id));

export function isAiFeature(value: string): value is AiFeatureId {
	return FEATURE_IDS.has(value);
}

export function featureLabel(id: string): string {
	return AI_FEATURES.find((feature) => feature.id === id)?.label ?? id;
}

let idSeq = 0;
export function uid(): string {
	idSeq += 1;
	return `m${idSeq.toString(36)}-${Math.random().toString(36).slice(2, 7)}`;
}

export type ModelDraft = { name: string; endpoint: string; timeout: string; maxContext: string };

export function parseModelDraft(form: FormData): ModelDraft | { error: string } {
	const name = String(form.get('name') ?? '')
		.replace(/\s+/g, ' ')
		.trim()
		.slice(0, 60);
	const endpoint = String(form.get('endpoint') ?? '').trim().slice(0, 200);
	const timeout = String(form.get('timeout') ?? '');
	const maxContext = String(form.get('maxContext') ?? '');
	if (!name) return { error: 'Give the model a name.' };
	if (!/^https?:\/\/\S+$/i.test(endpoint)) return { error: 'The endpoint must be an http(s) URL.' };
	return {
		name,
		endpoint,
		timeout: TIMEOUT_OPTIONS.includes(timeout) ? timeout : TIMEOUT_OPTIONS[1],
		maxContext: CONTEXT_OPTIONS.includes(maxContext) ? maxContext : CONTEXT_OPTIONS[1]
	};
}
