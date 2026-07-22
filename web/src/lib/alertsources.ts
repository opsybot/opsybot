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
	ingestUrl: string;
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

export const ADD_STEPS = ['Format', 'Endpoint', 'Mapping', 'Verify'];

export type GroupRule = { id: string; fields: string[]; windowSeconds: number };

export const GROUP_WINDOWS = [
	{ value: '300', label: '5 minutes' },
	{ value: '900', label: '15 minutes' },
	{ value: '3600', label: '1 hour' },
	{ value: '21600', label: '6 hours' },
	{ value: '86400', label: '24 hours' }
];

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

export const RT_SAMPLE = `{
  "title": "payments-api p99 above 800 ms",
  "severity": "high",
  "source": "prometheus-prod",
  "service": "payments-api",
  "labels": { "env": "prod", "team": "payments", "region": "eu-west-1" }
}`;
