import { SEVERITY_TONE as ALERT_TONE, SEVERITY_SHORT } from '$lib/alerts';
import { dependentsOf, type LinkKind, type Service } from '$lib/catalog';
import { SEVERITY_TONE } from '$lib/dashboard';
import { isActive } from '$lib/incidents';
import { listAlerts } from './alerts';
import { scenario } from './fixtures';
import { listIncidents } from './incidents';

const link = (runbook = '', dashboard = '', repository = ''): Record<LinkKind, string> => ({
	runbook,
	dashboard,
	repository
});

function seed(): Service[] {
	return [
		{
			id: 'payments-api',
			team: 'payments',
			description: 'Payment authorization, capture, and refunds. The money path.',
			links: link('runbooks/payments-api', 'grafana/payments', 'acme/payments-api'),
			deps: ['database', 'events-worker'],
			statusComponents: ['Checkout', 'Payments API']
		},
		{
			id: 'gateway',
			team: 'platform',
			description: 'Public API gateway: routing, auth, rate limits.',
			links: link('runbooks/gateway', 'grafana/gateway'),
			deps: ['edge'],
			statusComponents: ['Public API']
		},
		{
			id: 'checkout-web',
			team: 'frontend',
			description: 'Customer-facing checkout flow.',
			links: link('', 'grafana/checkout', 'acme/checkout-web'),
			deps: ['payments-api', 'gateway'],
			statusComponents: ['Checkout']
		},
		{
			id: 'database',
			team: 'platform',
			description: 'Primary Postgres cluster and replicas.',
			links: link('runbooks/database'),
			deps: [],
			statusComponents: []
		},
		{
			id: 'events-worker',
			team: 'payments',
			description: 'Async jobs: webhooks, notifications, ledger sync.',
			links: link('', '', 'acme/events-worker'),
			deps: ['database'],
			statusComponents: []
		},
		{
			id: 'edge',
			team: 'frontend',
			description: 'CDN, TLS termination, WAF.',
			links: link('', 'grafana/edge'),
			deps: [],
			statusComponents: ['Public API', 'Checkout']
		}
	];
}

const store = seed();
let empty = scenario() === 'empty';

function openAlerts(serviceId: string) {
	return listAlerts()
		.filter((alert) => alert.service === serviceId && alert.status !== 'resolved')
		.sort((a, b) => Date.parse(b.lastSeenAt) - Date.parse(a.lastSeenAt));
}

const THIRTY_DAYS = 30 * 24 * 3_600_000;

function incidentsOn(serviceId: string, sinceMs = 0) {
	return listIncidents()
		.filter(
			(incident) =>
				incident.services.includes(serviceId) &&
				(!sinceMs || Date.parse(incident.declaredAt) >= sinceMs)
		)
		.sort((a, b) => Date.parse(b.declaredAt) - Date.parse(a.declaredAt));
}

export type ServiceRow = {
	id: string;
	team: string;
	description: string;
	openAlerts: number;
	openIncidents: number;
	dependsOn: number;
	dependedOnBy: number;
};

export function listServices(): ServiceRow[] {
	if (empty) return [];

	return store.map((service) => ({
		id: service.id,
		team: service.team,
		description: service.description,
		openAlerts: openAlerts(service.id).length,
		openIncidents: incidentsOn(service.id).filter(isActive).length,
		dependsOn: service.deps.length,
		dependedOnBy: dependentsOf(service.id, store).length
	}));
}

export function getService(serviceId: string): Service | undefined {
	return empty ? undefined : store.find((service) => service.id === serviceId);
}

export function serviceNames(): string[] {
	return store.map((service) => service.id);
}

function dependents(serviceId: string): string[] {
	return dependentsOf(serviceId, store);
}

export function serviceActivity(serviceId: string) {
	return {
		openIncidents: incidentsOn(serviceId).filter(isActive).length,
		incidents: incidentsOn(serviceId, Date.now() - THIRTY_DAYS)
			.slice(0, 5)
			.map((incident) => ({
				id: incident.id,
				name: incident.name,
				severity: incident.severity,
				tone: SEVERITY_TONE[incident.severity],
				status: incident.status,
				active: isActive(incident)
			})),
		alerts: openAlerts(serviceId).map((alert) => ({
			id: alert.id,
			title: alert.title,
			severity: SEVERITY_SHORT[alert.severity],
			tone: ALERT_TONE[alert.severity],
			lastSeenAt: alert.lastSeenAt
		})),
		dependedOnBy: dependents(serviceId)
	};
}

function validDeps(deps: string[], selfNames: string[]): string[] {
	const known = new Set(store.map((service) => service.id));
	return [...new Set(deps)].filter((dep) => known.has(dep) && !selfNames.includes(dep));
}

export type ServiceInput = {
	name: string;
	team: string;
	description: string;
	links: Record<LinkKind, string>;
	deps: string[];
};

const RESERVED = ['new'];

export function nameTaken(name: string, exceptId?: string): boolean {
	if (RESERVED.includes(name)) return true;
	return store.some((service) => service.id === name && service.id !== exceptId);
}

export function createService(input: ServiceInput): Service {
	const service: Service = {
		id: input.name,
		team: input.team,
		description: input.description,
		links: input.links,
		deps: validDeps(input.deps, [input.name]),
		statusComponents: []
	};

	empty = false;
	store.push(service);
	return service;
}

export function updateService(serviceId: string, input: ServiceInput) {
	const service = getService(serviceId);
	if (!service) return;

	service.team = input.team;
	service.description = input.description;
	service.links = input.links;
	service.deps = validDeps(input.deps, [serviceId, input.name]);

	if (input.name !== serviceId) {
		service.id = input.name;
		for (const other of store) {
			other.deps = other.deps.map((dep) => (dep === serviceId ? input.name : dep));
		}
	}
}
