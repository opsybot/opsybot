import type { Platform } from '$lib/chat';
import { ANNOUNCE_CHANNELS, DEFAULT_DEFAULTS, PLATFORMS } from '$lib/chat';
import { scenario } from './fixtures';

function seed(): Platform[] {
	return PLATFORMS.map(
		(platform): Platform => ({
			...platform,
			connection:
				platform.id === 'slack'
					? {
							workspace: 'Acme Corp',
							health: 'healthy',
							healthNote: 'bot responded in 0.4 s · checked 2 m ago',
							defaults: { ...DEFAULT_DEFAULTS }
						}
					: null
		})
	);
}

const store = { platforms: seed() };

const state = scenario();
if (state === 'empty') {
	for (const platform of store.platforms) platform.connection = null;
}
if (state === 'degraded') {
	const slack = store.platforms.find((platform) => platform.id === 'slack');
	if (slack?.connection) {
		slack.connection.health = 'failing';
		slack.connection.healthNote = 'no response · last checked 6 m ago';
	}
}

export function listPlatforms(): Platform[] {
	return store.platforms;
}

function get(id: string): Platform | undefined {
	return store.platforms.find((platform) => platform.id === id);
}

export function sanitizeNamingPattern(input: string): string {
	const cleaned = input.replace(/\s+/g, ' ').trim().slice(0, 60);
	return cleaned || DEFAULT_DEFAULTS.namingPattern;
}

export function connect(id: string): boolean {
	const platform = get(id);
	if (!platform) return false;
	if (!platform.connection) {
		platform.connection = {
			workspace: 'Acme Corp',
			health: 'healthy',
			healthNote: 'bot responded in 0.5 s · checked just now',
			defaults: { ...DEFAULT_DEFAULTS }
		};
	}
	return true;
}

export function disconnect(id: string): boolean {
	const platform = get(id);
	if (!platform?.connection) return false;
	platform.connection = null;
	return true;
}

export function reconnect(id: string): boolean {
	const platform = get(id);
	if (!platform?.connection) return false;
	platform.connection.health = 'healthy';
	platform.connection.healthNote = 'scopes refreshed · checked just now';
	return true;
}

export function setNamingPattern(id: string, pattern: string): boolean {
	const platform = get(id);
	if (!platform?.connection) return false;
	platform.connection.defaults.namingPattern = sanitizeNamingPattern(pattern);
	return true;
}

export function setAnnounceChannel(id: string, channel: string): boolean {
	const platform = get(id);
	if (!platform?.connection || !ANNOUNCE_CHANNELS.includes(channel)) return false;
	platform.connection.defaults.announceChannel = channel;
	return true;
}

export function setArchiveOnResolve(id: string, archive: boolean): boolean {
	const platform = get(id);
	if (!platform?.connection) return false;
	platform.connection.defaults.archiveOnResolve = archive;
	return true;
}
