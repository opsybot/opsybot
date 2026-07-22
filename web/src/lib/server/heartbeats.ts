import type { Heartbeat } from '$lib/alerts';
import { scenario } from './fixtures';

const MINUTE = 60 * 1000;
const HOUR = 60 * MINUTE;
const iso = (offsetMs: number) => new Date(Date.now() + offsetMs).toISOString();
const id = (prefix: string) => prefix + Math.random().toString(36).slice(2, 8);
const spaced = (duration: string) => duration.replace(/^(\d+)/, '$1 ');

function seed(): Heartbeat[] {
	return [
		{
			id: 'hb-1',
			name: 'nightly-db-backup',
			state: 'healthy',
			interval: '24 h',
			grace: '30 m',
			lastSeenAt: iso(-7 * HOUR),
			policy: 'platform-default'
		},
		{
			id: 'hb-2',
			name: 'billing-cron',
			state: 'missed',
			interval: '1 h',
			grace: '10 m',
			lastSeenAt: iso(-2 * HOUR - 14 * MINUTE),
			policy: 'payments-primary'
		},
		{
			id: 'hb-3',
			name: 'edge-sync',
			state: 'healthy',
			interval: '5 m',
			grace: '2 m',
			lastSeenAt: iso(-42_000),
			policy: 'platform-default'
		}
	]
}

const store = seed();
const empty = scenario() === 'empty';

export function listHeartbeats(): Heartbeat[] {
	return empty ? [] : store;
}

export function createHeartbeat(input: {
	name: string;
	interval: string;
	grace: string;
	policy: string;
}): { slug: string; token: string } {
	const slug = input.name
		.trim()
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');

	store.push({
		id: id('hb-'),
		name: input.name.trim(),
		state: 'healthy',
		interval: spaced(input.interval),
		grace: spaced(input.grace),
		lastSeenAt: null,
		policy: input.policy
	});

	return { slug, token: Math.random().toString(16).slice(2, 8) };
}
