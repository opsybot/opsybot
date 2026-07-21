import type { Condition, DeliveryEvent, Mapping, RoutingRule, Source } from '$lib/alertsources';
import { FORMATS, MAPPINGS, mappingKeyFor, slugify } from '$lib/alertsources';
import { scenario } from './fixtures';

const secret = (seed: string) => `osk_${seed}`;

function seed() {
	const sources: Source[] = [
		{
			id: 'prometheus-prod', name: 'prometheus-prod', format: 'Prometheus Alertmanager', icon: 'flame',
			health: 'healthy', lastEvent: '2 m ago', status: 'active', slug: 'am-prod',
			volume: [4, 6, 3, 5, 8, 7, 4, 6, 9, 12, 8, 6, 5, 7, 6, 8, 10, 7, 5, 6, 8, 9, 7, 6],
			failures: 0, secret: secret('9f27c3a1e8b44d17'), mapping: MAPPINGS.alertmanager.map((m) => ({ ...m }))
		},
		{
			id: 'grafana-main', name: 'grafana-main', format: 'Grafana', icon: 'chart-line',
			health: 'healthy', lastEvent: '18 m ago', status: 'active', slug: 'grafana',
			volume: [2, 1, 3, 2, 4, 3, 2, 1, 2, 3, 5, 4, 2, 3, 2, 4, 3, 2, 3, 4, 2, 3, 2, 1],
			failures: 0, secret: secret('b41d09e7c2a86f30'), mapping: MAPPINGS.grafana.map((m) => ({ ...m }))
		},
		{
			id: 'uptime-edge', name: 'uptime-edge', format: 'Uptime Kuma', icon: 'globe',
			health: 'stale', lastEvent: '3 d ago', status: 'active', slug: 'kuma-edge',
			volume: [1, 0, 2, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
			failures: 0, secret: secret('7c3e11a9d0554b62'), mapping: MAPPINGS.kuma.map((m) => ({ ...m }))
		},
		{
			id: 'legacy-nagios', name: 'legacy-nagios', format: 'Generic JSON', icon: 'braces',
			health: 'failing', lastEvent: '1 h ago', status: 'active', slug: 'nagios',
			volume: [3, 2, 4, 3, 2, 3, 4, 2, 3, 2, 3, 4, 3, 2, 3, 4, 2, 3, 4, 3, 2, 3, 2, 3],
			failures: 3, secret: secret('2a90f4bc6e1d7738'), mapping: MAPPINGS.generic.map((m) => ({ ...m }))
		},
		{
			id: 'billing-cron', name: 'billing-cron', format: 'Heartbeat', icon: 'heart-pulse',
			health: 'healthy', lastEvent: '42 s ago', status: 'paused', slug: 'hb-billing',
			volume: [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
			failures: 0, secret: secret('e5b8207ad9346c11'), mapping: MAPPINGS.heartbeat.map((m) => ({ ...m }))
		}
	];

	const rules: RoutingRule[] = [
		{ id: 'r1', conditions: [{ field: 'service', op: 'is', value: 'payments-api' }], policy: 'payments-primary' },
		{ id: 'r2', conditions: [{ field: 'labels.env', op: 'is', value: 'staging' }], policy: 'frontend-daytime' },
		{
			id: 'r3',
			conditions: [
				{ field: 'severity', op: 'is', value: 'critical' },
				{ field: 'labels.region', op: 'matches', value: 'eu-*' }
			],
			policy: 'payments-primary'
		}
	];

	return { sources, rules, defaultPolicy: 'platform-default' };
}

const store = seed();
const empty = scenario() === 'empty';
if (empty) {
	store.sources = [];
	store.rules = [];
}

const EVENTS: DeliveryEvent[] = [
	{ at: '2026-07-11 09:48:11 UTC', title: 'payments-api p99 above 800 ms', outcome: 'deduped ×12', tone: 'neutral' },
	{ at: '2026-07-11 09:41:03 UTC', title: 'Disk usage 85% on db-3', outcome: 'alert created', tone: 'success' },
	{ at: '2026-07-11 08:02:14 UTC', title: 'unparseable', outcome: 'parse failed', tone: 'critical' },
	{ at: '2026-07-11 06:40:52 UTC', title: 'payments-api p99 above 800 ms', outcome: 'alert created', tone: 'success' },
	{ at: '2026-07-10 23:19:40 UTC', title: 'Node exporter down on worker-7', outcome: 'alert created', tone: 'success' }
];

function id16(): string {
	let hex = '';
	while (hex.length < 16) hex += Math.floor(Math.random() * 16).toString(16);
	return hex.slice(0, 16);
}

const RESERVED = new Set(['new', 'routing']);

export function sanitizeMapping(raw: string): Mapping[] {
	let data: unknown;
	try {
		data = JSON.parse(raw);
	} catch {
		return [];
	}
	if (!Array.isArray(data)) return [];
	return data
		.filter((row): row is Record<string, unknown> => !!row && typeof row === 'object' && typeof row.field === 'string')
		.map((row) => ({ field: (row.field as string).trim(), path: typeof row.path === 'string' ? row.path : '' }))
		.filter((row) => row.field)
		.filter((row, index, all) => all.findIndex((other) => other.field === row.field) === index);
}

export function listSources(): Source[] {
	return store.sources;
}

export function getSource(id: string): Source | undefined {
	return store.sources.find((source) => source.id === id);
}

export function eventsFor(source: Source): DeliveryEvent[] {
	return source.failures ? EVENTS : EVENTS.filter((event) => event.outcome !== 'parse failed');
}

export function draftSecret(): string {
	return secret(id16());
}

export function createSource(input: {
	name: string;
	formatId: string;
	mapping: Mapping[];
	secret: string;
}): Source {
	const format = FORMATS.find((entry) => entry.id === input.formatId) ?? FORMATS[FORMATS.length - 1];
	const base = slugify(input.name) || format.id;
	let unique = base;
	for (
		let n = 2;
		RESERVED.has(unique) || store.sources.some((source) => source.id === unique || source.slug === unique);
		n++
	) {
		unique = `${base}-${n}`;
	}

	const source: Source = {
		id: unique,
		name: input.name.trim(),
		format: format.label,
		icon: format.icon,
		health: 'healthy',
		lastEvent: 'just now',
		status: 'active',
		slug: unique,
		volume: Array(24).fill(0),
		failures: 0,
		secret: input.secret,
		mapping: input.mapping.length ? input.mapping : MAPPINGS[format.id].map((m) => ({ ...m }))
	};
	store.sources.push(source);
	return source;
}

export function setPaused(id: string, paused: boolean): boolean {
	const source = getSource(id);
	if (!source) return false;
	source.status = paused ? 'paused' : 'active';
	return true;
}

export function rotateSecret(id: string): string | undefined {
	const source = getSource(id);
	if (!source) return undefined;
	source.secret = secret(id16());
	return source.secret;
}

export function saveMapping(id: string, mapping: Mapping[]): boolean {
	const source = getSource(id);
	if (!source) return false;
	source.mapping = mapping;
	return true;
}

export function listRules(): RoutingRule[] {
	return store.rules;
}

export function defaultPolicy(): string {
	return store.defaultPolicy;
}

export function setDefaultPolicy(policy: string) {
	store.defaultPolicy = policy;
}

export function addRule(rule: { conditions: Condition[]; policy: string }, position: string) {
	const created: RoutingRule = { id: `r${id16().slice(0, 6)}`, conditions: rule.conditions, policy: rule.policy };
	if (position === 'start') {
		store.rules.unshift(created);
	} else if (position === 'end') {
		store.rules.push(created);
	} else {
		const at = Number.parseInt(position, 10);
		const index = Number.isFinite(at) ? Math.min(Math.max(at, 0), store.rules.length) : store.rules.length;
		store.rules.splice(index, 0, created);
	}
}

export function updateRule(id: string, rule: { conditions: Condition[]; policy: string }): boolean {
	const existing = store.rules.find((entry) => entry.id === id);
	if (!existing) return false;
	existing.conditions = rule.conditions;
	existing.policy = rule.policy;
	return true;
}

export function deleteRule(id: string) {
	store.rules = store.rules.filter((rule) => rule.id !== id);
}

export function moveRule(id: string, by: -1 | 1) {
	const from = store.rules.findIndex((rule) => rule.id === id);
	const to = from + by;
	if (from < 0 || to < 0 || to >= store.rules.length) return;
	[store.rules[from], store.rules[to]] = [store.rules[to], store.rules[from]];
}
