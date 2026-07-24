import type { Cookies } from '@sveltejs/kit';
import type { Alert } from '$lib/alerts';
import { SEVERITY_TONE as ALERT_TONE, SEVERITY_SHORT } from '$lib/alerts';
import type { Service } from '$lib/catalog';
import { SEVERITY_TONE } from '$lib/dashboard';
import { isActive } from '$lib/incidents';
import { listIncidents } from './incidents-api';
import {
	createService as apiCreateService,
	deleteService as apiDeleteService,
	listServices as apiListServices,
	updateService as apiUpdateService,
	type CatalogService
} from './services-api';

function shape(service: CatalogService): Service {
	return {
		id: service.slug,
		team: service.team,
		description: service.description,
		links: { runbook: '', dashboard: '', repository: '' },
		deps: [],
		statusComponents: []
	};
}

function openAlerts(alerts: Alert[], serviceId: string) {
	return alerts
		.filter((alert) => alert.service === serviceId && alert.status !== 'resolved')
		.sort((a, b) => Date.parse(b.lastSeenAt) - Date.parse(a.lastSeenAt));
}

const THIRTY_DAYS = 30 * 24 * 3_600_000;

export type ServiceRow = {
	id: string;
	team: string;
	description: string;
	openAlerts: number;
	openIncidents: number;
	dependsOn: number;
	dependedOnBy: number;
};

export async function listServices(
	cookies: Cookies,
	workspace: string,
	alerts: Alert[] = []
): Promise<ServiceRow[]> {
	const [services, incidentsPage] = await Promise.all([
		apiListServices(cookies, workspace),
		listIncidents(cookies, workspace, { limit: 100 })
	]);
	const incidents = incidentsPage.incidents;
	return services.map((service) => ({
		id: service.slug,
		team: service.team,
		description: service.description,
		openAlerts: openAlerts(alerts, service.slug).length,
		openIncidents: incidents.filter(
			(incident) => incident.services.includes(service.name) && isActive(incident)
		).length,
		dependsOn: 0,
		dependedOnBy: 0
	}));
}

async function findBySlug(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<CatalogService | undefined> {
	const services = await apiListServices(cookies, workspace);
	return services.find((service) => service.slug === slug);
}

export async function getService(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<Service | undefined> {
	const service = await findBySlug(cookies, workspace, slug);
	return service ? shape(service) : undefined;
}

export async function serviceNames(cookies: Cookies, workspace: string): Promise<string[]> {
	const services = await apiListServices(cookies, workspace);
	return services.map((service) => service.slug);
}

export async function serviceActivity(
	cookies: Cookies,
	workspace: string,
	slug: string,
	alerts: Alert[] = []
) {
	const service = await findBySlug(cookies, workspace, slug);
	const name = service?.name ?? slug;
	const { incidents } = await listIncidents(cookies, workspace, { limit: 100 });
	const onService = incidents
		.filter((incident) => incident.services.includes(name))
		.sort((a, b) => Date.parse(b.declaredAt) - Date.parse(a.declaredAt));
	const since = Date.now() - THIRTY_DAYS;

	return {
		openIncidents: onService.filter(isActive).length,
		incidents: onService
			.filter((incident) => Date.parse(incident.declaredAt) >= since)
			.slice(0, 5)
			.map((incident) => ({
				id: incident.id,
				name: incident.name,
				severity: incident.severity,
				tone: SEVERITY_TONE[incident.severity],
				status: incident.status,
				active: isActive(incident)
			})),
		alerts: openAlerts(alerts, slug).map((alert) => ({
			id: alert.id,
			title: alert.title,
			severity: SEVERITY_SHORT[alert.severity],
			tone: ALERT_TONE[alert.severity],
			lastSeenAt: alert.lastSeenAt
		})),
		dependedOnBy: [] as string[]
	};
}

export type ServiceInput = {
	name: string;
	team: string;
	description: string;
};

export async function createService(
	cookies: Cookies,
	workspace: string,
	input: ServiceInput
): Promise<{ slug?: string; error?: string }> {
	const result = await apiCreateService(cookies, workspace, {
		name: input.name,
		team: input.team,
		description: input.description
	});
	if (result.error || !result.id) return { error: result.error ?? 'Could not create the service.' };
	const services = await apiListServices(cookies, workspace);
	const created = services.find((service) => service.id === result.id);
	return { slug: created?.slug };
}

export async function updateService(
	cookies: Cookies,
	workspace: string,
	slug: string,
	input: ServiceInput
): Promise<{ slug?: string; error?: string }> {
	const service = await findBySlug(cookies, workspace, slug);
	if (!service) return { error: 'That service no longer exists.' };
	const result = await apiUpdateService(cookies, workspace, service.id, {
		name: input.name,
		team: input.team,
		description: input.description
	});
	if (result.error) return { error: result.error };
	return { slug };
}

export async function removeService(
	cookies: Cookies,
	workspace: string,
	slug: string
): Promise<{ error?: string }> {
	const service = await findBySlug(cookies, workspace, slug);
	if (!service) return { error: 'That service no longer exists.' };
	return apiDeleteService(cookies, workspace, service.id);
}
