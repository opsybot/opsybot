import type { Tone } from '$lib/dashboard';

export type PlatformId = 'slack' | 'teams' | 'discord';
export type Health = 'healthy' | 'failing';

export type Scope = { what: string; why: string };

export type ChannelDefaults = {
	namingPattern: string;
	announceChannel: string;
	archiveOnResolve: boolean;
};

export type Connection = {
	workspace: string;
	health: Health;
	healthNote: string;
	defaults: ChannelDefaults;
};

export type Platform = {
	id: PlatformId;
	label: string;
	icon: string;
	tagline: string;
	scopes: Scope[];
	connection: Connection | null;
};

export const INSTALL_STEPS = ['consent', 'waiting', 'done', 'tested'] as const;
export type InstallStep = (typeof INSTALL_STEPS)[number];

export const ANNOUNCE_CHANNELS = ['#incidents', '#eng-all', '#ops'];

export const DEFAULT_DEFAULTS: ChannelDefaults = {
	namingPattern: 'inc-{number}',
	announceChannel: '#incidents',
	archiveOnResolve: true
};

export const PLATFORMS: Omit<Platform, 'connection'>[] = [
	{
		id: 'slack',
		label: 'Slack',
		icon: 'message-square',
		tagline: 'Incidents run in chat. Declare, coordinate, resolve without leaving Slack.',
		scopes: [
			{ what: 'Create and archive channels', why: 'to open #inc-NNNN rooms and close them on resolve' },
			{ what: 'Post and read in incident channels', why: 'the timeline scribe works from channel messages' },
			{ what: 'Send DMs', why: 'pages and personal notifications' },
			{ what: 'Look up members by email', why: 'to match responders to Slack accounts' }
		]
	},
	{
		id: 'teams',
		label: 'Microsoft Teams',
		icon: 'message-square',
		tagline: 'Incidents run in chat. Declare, coordinate, resolve without leaving Microsoft Teams.',
		scopes: [
			{ what: 'Create channels in a team you pick', why: 'incident rooms live in one team' },
			{ what: 'Post and read in incident channels', why: 'the timeline scribe works from channel messages' },
			{ what: 'Send chats', why: 'pages and personal notifications' }
		]
	},
	{
		id: 'discord',
		label: 'Discord',
		icon: 'message-square',
		tagline: 'Incidents run in chat. Declare, coordinate, resolve without leaving Discord.',
		scopes: [
			{ what: 'Manage channels in one category', why: 'incident rooms are created under it' },
			{ what: 'Post and read in incident channels', why: 'the timeline scribe works from channel messages' },
			{ what: 'Send DMs', why: 'pages and personal notifications' }
		]
	}
];

export function connectionBadge(platform: Pick<Platform, 'connection'>): { tone: Tone; label: string; dot: boolean } {
	const connection = platform.connection;
	if (!connection) return { tone: 'neutral', label: 'not connected', dot: false };
	if (connection.health === 'failing') return { tone: 'critical', label: 'not responding', dot: true };
	return { tone: 'success', label: 'connected', dot: true };
}

export function previewChannelName(
	pattern: string,
	ctx: { number: number; slug: string } = { number: 2481, slug: 'checkout-degraded' }
): string {
	return pattern.replaceAll('{number}', String(ctx.number)).replaceAll('{slug}', ctx.slug);
}
