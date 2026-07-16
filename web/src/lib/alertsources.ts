import type { Tone } from '$lib/dashboard';

export type Health = 'healthy' | 'stale' | 'failing';
export type SourceStatus = 'active' | 'paused';

export type Source = {
	id: string;
	name: string;
	format: string;
	icon: string;
	health: Health;
	lastEvent: string;
	status: SourceStatus;
	slug: string;
	volume: number[];
	failures: number;
	secret: string;
	mapping: Mapping[];
};

export type Mapping = { field: string; path: string };

export type Format = {
	id: string;
	label: string;
	icon: string;
	desc: string;
};

export const FORMATS: Format[] = [
	{ id: 'alertmanager', label: 'Prometheus Alertmanager', icon: 'flame', desc: 'Webhook receiver for Alertmanager routes' },
	{ id: 'grafana', label: 'Grafana', icon: 'chart-line', desc: 'Grafana alerting contact point' },
	{ id: 'kuma', label: 'Uptime Kuma', icon: 'globe', desc: 'Uptime Kuma webhook notifications' },
	{ id: 'heartbeat', label: 'Heartbeat', icon: 'heart-pulse', desc: 'Pages when a job stops checking in' },
	{ id: 'generic', label: 'Generic JSON', icon: 'braces', desc: 'Any tool that can POST JSON' }
];

export const MAPPINGS: Record<string, Mapping[]> = {
	alertmanager: [
		{ field: 'title', path: 'alerts[0].annotations.summary' },
		{ field: 'severity', path: 'alerts[0].labels.severity' },
		{ field: 'service', path: 'alerts[0].labels.service' },
		{ field: 'labels', path: 'alerts[0].labels.*' }
	],
	grafana: [
		{ field: 'title', path: 'title' },
		{ field: 'severity', path: 'labels.severity' },
		{ field: 'service', path: 'labels.service' },
		{ field: 'labels', path: 'labels.*' }
	],
	kuma: [
		{ field: 'title', path: 'msg' },
		{ field: 'severity', path: '"warning" (fixed)' },
		{ field: 'service', path: 'monitor.name' },
		{ field: 'labels', path: 'monitor.tags.*' }
	],
	heartbeat: [
		{ field: 'title', path: '"<name> missed check-in" (generated)' },
		{ field: 'severity', path: 'monitor grace config' },
		{ field: 'service', path: 'monitor routing' }
	],
	generic: [
		{ field: 'title', path: 'title' },
		{ field: 'severity', path: 'severity' },
		{ field: 'service', path: 'service' },
		{ field: 'labels', path: 'labels.*' }
	]
};

export function mappingKeyFor(formatLabel: string): string {
	return FORMATS.find((format) => format.label === formatLabel)?.id ?? 'generic';
}

export const ADD_STEPS = ['Format', 'Endpoint', 'Mapping', 'Verify'];

export type DeliveryEvent = {
	at: string;
	title: string;
	outcome: string;
	tone: 'neutral' | 'success' | 'critical';
};

export function endpointUrl(slug: string): string {
	return `https://in.opsy.bot/e/acme/${slug}`;
}

export function slugify(name: string): string {
	return name
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '');
}

export function healthBadge(source: Pick<Source, 'health' | 'failures'>): { tone: Tone; label: string } {
	if (source.health === 'failing') return { tone: 'critical', label: `${source.failures} parse failures` };
	if (source.health === 'stale') return { tone: 'warning', label: 'no events · 3 d' };
	return { tone: 'success', label: 'healthy' };
}

export type ConditionOp = 'is' | 'is not' | 'contains' | 'matches';
export type Condition = { field: string; op: ConditionOp; value: string };
export type RoutingRule = { id: string; conditions: Condition[]; policy: string };

export const RT_FIELDS = ['source', 'service', 'severity', 'title', 'labels.env', 'labels.team', 'labels.region'];
export const RT_OPS: ConditionOp[] = ['is', 'is not', 'contains', 'matches'];
export const RT_POLICIES = ['payments-primary', 'platform-default', 'frontend-daytime'];

export const RT_SAMPLE = `{
  "title": "payments-api p99 above 800 ms",
  "severity": "high",
  "source": "prometheus-prod",
  "service": "payments-api",
  "labels": { "env": "prod", "team": "payments", "region": "eu-west-1" }
}`;

function getPath(object: unknown, path: string): unknown {
	return path.split('.').reduce<unknown>((value, key) => {
		if (value == null || typeof value !== 'object') return undefined;
		return (value as Record<string, unknown>)[key];
	}, object);
}

// Linear-scan glob; no regex, so adversarial patterns cannot cause backtracking blowup
function globMatch(value: string, pattern: string): boolean {
	const haystack = value.toLowerCase();
	const parts = pattern.toLowerCase().split('*');
	if (parts.length === 1) return haystack === parts[0];

	const first = parts[0];
	const last = parts[parts.length - 1];
	if (!haystack.startsWith(first) || !haystack.endsWith(last)) return false;

	let cursor = first.length;
	for (let i = 1; i < parts.length - 1; i++) {
		const part = parts[i];
		if (!part) continue;
		const at = haystack.indexOf(part, cursor);
		if (at < 0) return false;
		cursor = at + part.length;
	}
	// The prefix and middles must not overrun into the suffix region
	return cursor <= haystack.length - last.length;
}

export function conditionMatches(alert: unknown, condition: Condition): boolean {
	const raw = getPath(alert, condition.field);
	const value = raw == null ? '' : String(raw);
	switch (condition.op) {
		case 'is':
			return value === condition.value;
		case 'is not':
			return value !== condition.value;
		case 'contains':
			return value.toLowerCase().includes(condition.value.toLowerCase());
		case 'matches':
			return globMatch(value, condition.value);
	}
}

// Rules with no conditions never match; [].every() is true
export function matchRule(alert: unknown, rules: RoutingRule[]): number {
	for (let index = 0; index < rules.length; index++) {
		const rule = rules[index];
		if (rule.conditions.length && rule.conditions.every((condition) => conditionMatches(alert, condition))) {
			return index;
		}
	}
	return -1;
}

export function evaluateSample(
	sample: string,
	rules: RoutingRule[],
	defaultPolicy: string
): { ok: true; index: number; policy: string } | { ok: false; error: string } {
	let alert: unknown;
	try {
		alert = JSON.parse(sample);
	} catch (error) {
		return { ok: false, error: `Invalid JSON — ${(error as Error).message}` };
	}
	const index = matchRule(alert, rules);
	return { ok: true, index, policy: index === -1 ? defaultPolicy : rules[index].policy };
}
