import { isActive } from '$lib/incidents';
import type { Component, ComponentState, StatusPage, Visibility } from '$lib/statuspages';
import { STAGE_CAPITAL, type PublishStage } from '$lib/statuspages';
import { getIncident, listIncidents, publishStatusUpdate } from './incidents';
import { scenario } from './fixtures';

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

const id = (prefix: string) => prefix + Math.random().toString(36).slice(2, 8);
const iso = (offsetMs: number) => new Date(Date.now() + offsetMs).toISOString();

export type Maintenance = {
	id: string;
	title: string;
	description: string;
	components: string[];
	startsAt: string;
	endsAt: string;
	notice: string;
};

export type Subscriber = { address: string; since: string };
export type WebhookSubscriber = { url: string; ok: boolean; lastAt: string; lastResult: string };

function seed() {
	const components: Component[] = [
		{ id: 'c1', name: 'Checkout', group: 'Customer', services: ['checkout-web', 'payments-api'], state: 'degraded' },
		{ id: 'c2', name: 'Payments API', group: 'Customer', services: ['payments-api'], state: 'degraded' },
		{ id: 'c3', name: 'Public API', group: 'Platform', services: ['gateway', 'edge'], state: 'operational' },
		{ id: 'c4', name: 'Dashboard', group: 'Customer', services: [], state: 'operational' }
	];

	const pages: StatusPage[] = [
		{
			id: 'status.acme.dev',
			name: 'Acme status',
			description: 'Live status for acme.dev products.',
			pageTitle: 'Acme status: live',
			visibility: 'public',
			domain: 'status.acme.dev',
			domainVerified: true,
			certRenews: '2026-09-04',
			accent: 'mint',
			utcDefault: true,
			showUptime: true,
			allowIndexing: true,
			token: '9f27c3a1e8',
			published: true,
			components,
			subscribers: { email: 1211, feed: 58, webhook: 15 }
		},
		{
			id: 'internal.acme.dev',
			name: 'Internal platform status',
			description: 'Platform health for Acme engineers.',
			pageTitle: 'Internal platform status',
			visibility: 'auth',
			domain: 'internal.acme.dev',
			domainVerified: true,
			certRenews: '2026-08-22',
			accent: 'blue',
			utcDefault: true,
			showUptime: true,
			allowIndexing: false,
			token: 'b41d09e7c2',
			published: true,
			components: [
				{ id: 'c5', name: 'Build pipeline', group: 'Platform', services: [], state: 'operational' },
				{ id: 'c6', name: 'Internal API', group: 'Platform', services: ['gateway'], state: 'operational' }
			],
			subscribers: { email: 84, feed: 8, webhook: 4 }
		}
	];

	const maintenance: Maintenance[] = [
		{
			id: 'm1',
			title: 'Database maintenance: primary failover test',
			description: 'Brief interruptions to checkout while the primary fails over.',
			components: ['Payments API', 'Checkout'],
			startsAt: iso(DAY + 12 * HOUR),
			endsAt: iso(DAY + 16 * HOUR),
			notice: '48 h + 1 h before'
		},
		{
			id: 'm2',
			title: 'CDN migration',
			description: 'Moving edge traffic to the new CDN provider.',
			components: ['Public API'],
			startsAt: iso(-11 * DAY),
			endsAt: iso(-11 * DAY + 2 * HOUR),
			notice: '72 h before'
		},
		{
			id: 'm3',
			title: 'TLS certificate rotation',
			description: 'Rotating the edge certificate; no expected impact.',
			components: ['Public API', 'Checkout'],
			startsAt: iso(-15 * DAY),
			endsAt: iso(-15 * DAY + 40 * MINUTE),
			notice: '48 h before'
		}
	];

	const emails: Subscriber[] = [
		{ address: 'ops@bigcustomer.io', since: '2026-03-14' },
		{ address: 'sre-team@fintechco.com', since: '2026-04-02' },
		{ address: 'alerts@shopfast.dev', since: '2026-05-19' },
		{ address: 'noc@enterprise-corp.com', since: '2026-06-01' },
		{ address: 'platform@startuplab.io', since: '2026-06-28' }
	];

	const webhooks: WebhookSubscriber[] = [
		{
			url: 'https://hooks.bigcustomer.io/status',
			ok: true,
			lastAt: iso(-11 * MINUTE),
			lastResult: '200 · 0.3 s'
		},
		{
			url: 'https://fintechco.com/api/vendor-status',
			ok: false,
			lastAt: iso(-11 * MINUTE),
			lastResult: '503 · retried ×3'
		}
	];

	return { pages, maintenance, emails, webhooks };
}

const store = seed();
let empty = scenario() === 'empty';

export function listPages(): StatusPage[] {
	return empty ? [] : store.pages;
}

export function getPage(pageId: string): StatusPage | undefined {
	return empty ? undefined : store.pages.find((page) => page.id === pageId);
}

const RESERVED = ['maintenance', 'subscribers', 'publish', 'new'];

export function pageNameTaken(domain: string, exceptId?: string): boolean {
	if (RESERVED.includes(domain)) return true;
	return store.pages.some((page) => page.id === domain && page.id !== exceptId);
}

export type PageSettings = {
	name: string;
	description: string;
	pageTitle: string;
	domain: string;
	visibility: Visibility;
	accent: string;
	utcDefault: boolean;
	showUptime: boolean;
	allowIndexing: boolean;
};

export function updatePage(pageId: string, settings: PageSettings) {
	const page = getPage(pageId);
	if (!page) return;

	Object.assign(page, {
		name: settings.name,
		description: settings.description,
		pageTitle: settings.pageTitle,
		visibility: settings.visibility,
		accent: settings.accent,
		utcDefault: settings.utcDefault,
		showUptime: settings.showUptime,
		allowIndexing: settings.allowIndexing
	});

	if (settings.domain !== pageId) {
		page.id = settings.domain;
		page.domainVerified = false;
	}
	page.domain = settings.domain;
}

export function moveComponent(pageId: string, componentId: string, by: -1 | 1) {
	const page = getPage(pageId);
	if (!page) return;

	const from = page.components.findIndex((component) => component.id === componentId);
	const to = from + by;
	if (from < 0 || to < 0 || to >= page.components.length) return;

	[page.components[from], page.components[to]] = [page.components[to], page.components[from]];
}

export function setPublished(pageId: string, published: boolean) {
	const page = getPage(pageId);
	if (page) page.published = published;
}

export function rotateToken(pageId: string): string | undefined {
	const page = getPage(pageId);
	if (!page) return;
	page.token = Math.random().toString(16).slice(2, 12);
	return page.token;
}

export function deletePage(pageId: string) {
	const index = store.pages.findIndex((page) => page.id === pageId);
	if (index >= 0) store.pages.splice(index, 1);
}

export function listMaintenance(now = Date.now()) {
	const sorted = [...store.maintenance].sort((a, b) => Date.parse(a.startsAt) - Date.parse(b.startsAt));
	return {
		upcoming: sorted.filter((window) => Date.parse(window.endsAt) >= now),
		past: sorted.filter((window) => Date.parse(window.endsAt) < now).reverse()
	};
}

export function scheduleMaintenance(input: {
	title: string;
	description: string;
	components: string[];
	startsAt: string;
	endsAt: string;
	notice: string;
}) {
	store.maintenance.push({ id: id('m'), ...input });
}

export function allComponentNames(): string[] {
	return [...new Set(store.pages.flatMap((page) => page.components.map((component) => component.name)))];
}

export function subscriberCounts() {
	const primary = store.pages.find((page) => page.visibility === 'public') ?? store.pages[0];
	return primary?.subscribers ?? { email: 0, feed: 0, webhook: 0 };
}

export function listEmails(query = ''): Subscriber[] {
	const needle = query.trim().toLowerCase();
	return needle
		? store.emails.filter((subscriber) => subscriber.address.toLowerCase().includes(needle))
		: store.emails;
}

export function removeEmail(address: string) {
	store.emails = store.emails.filter((subscriber) => subscriber.address !== address);
}

export function listWebhooks(): WebhookSubscriber[] {
	return store.webhooks;
}

export function redeliverWebhook(url: string) {
	const webhook = store.webhooks.find((entry) => entry.url === url);
	if (webhook) {
		webhook.ok = true;
		webhook.lastAt = new Date().toISOString();
		webhook.lastResult = '200 · 0.4 s';
	}
}

export function publishIncident(input: {
	incidentId: string;
	pageIds: string[];
	componentStates: Record<string, ComponentState>;
	title: string;
	stage: PublishStage;
	text: string;
}): boolean {
	if (!getIncident(input.incidentId)) return false;

	for (const pageId of input.pageIds) {
		const page = getPage(pageId);
		if (!page) continue;
		for (const component of page.components) {
			const next = input.componentStates[component.name];
			if (next) component.state = next;
		}
	}

	publishStatusUpdate(input.incidentId, STAGE_CAPITAL[input.stage], input.text, {
		domain: store.pages.find((page) => input.pageIds.includes(page.id))?.domain,
		title: input.title
	});
	return true;
}

export function postStatusUpdate(incidentId: string, stage: PublishStage, text: string) {
	publishStatusUpdate(incidentId, STAGE_CAPITAL[stage], text);
}

export function publishTarget(incidentId?: string) {
	if (incidentId) return getIncident(incidentId);

	return listIncidents()
		.filter(isActive)
		.sort((a, b) => Date.parse(b.declaredAt) - Date.parse(a.declaredAt))[0];
}
