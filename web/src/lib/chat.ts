import type { Tone } from '$lib/dashboard';

export type PlatformId = 'slack' | 'teams' | 'discord' | 'telegram';
export type Health = 'healthy' | 'failing';

export type Scope = { what: string; why: string };

export type AuthKind = 'oauth' | 'bot-token' | 'telegram';

export type ExternalIdField = { label: string; placeholder: string; hint: string };

export type ChannelDefaults = {
	namingPattern: string;
	announceChannel: string;
	archiveOnResolve: boolean;
};

export type LinkMethod = 'email' | 'oauth' | 'telegram';

export type Connection = {
	workspace: string;
	health: Health;
	healthNote: string;
	defaults: ChannelDefaults;
	linked: boolean;
	linkedHandle: string;
	linkedVerified: boolean;
	linkMethod?: LinkMethod;
};

export type Platform = {
	id: PlatformId;
	label: string;
	icon: string;
	tagline: string;
	authKind: AuthKind;
	scopes: Scope[];
	externalIdField?: ExternalIdField;
	connection: Connection | null;
};

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
		authKind: 'oauth',
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
		tagline: 'Get paged in Microsoft Teams and acknowledge without leaving the chat.',
		authKind: 'bot-token',
		scopes: [
			{ what: 'Message you directly', why: 'your pages and notifications arrive as a Teams chat' },
			{ what: 'Show Acknowledge / Resolve buttons', why: 'act on a page from the message' }
		]
	},
	{
		id: 'discord',
		label: 'Discord',
		icon: 'message-square',
		tagline: 'Incidents run in chat. Declare, coordinate, resolve without leaving Discord.',
		authKind: 'oauth',
		scopes: [
			{ what: 'Manage channels in one category', why: 'incident rooms are created under it' },
			{ what: 'Post and read in incident channels', why: 'the timeline scribe works from channel messages' },
			{ what: 'Send DMs', why: 'pages and personal notifications' }
		]
	},
	{
		id: 'telegram',
		label: 'Telegram',
		icon: 'message-square',
		tagline: 'Get paged in Telegram and acknowledge without leaving the chat.',
		authKind: 'telegram',
		scopes: [
			{ what: 'Message you directly', why: 'your pages and notifications arrive as a DM' },
			{ what: 'Show Acknowledge / Resolve buttons', why: 'act on a page from the message' }
		]
	}
];

export function isPlatformId(value: string): value is PlatformId {
	return PLATFORMS.some((platform) => platform.id === value);
}

export function connectionBadge(platform: Pick<Platform, 'connection'>): { tone: Tone; label: string; dot: boolean } {
	const connection = platform.connection;
	if (!connection) return { tone: 'neutral', label: 'not connected', dot: false };
	if (connection.health === 'failing') return { tone: 'critical', label: 'not responding', dot: true };
	return { tone: 'success', label: 'connected', dot: true };
}

export function linkBadge(connection: Connection | null): { tone: Tone; label: string; dot: boolean } {
	if (!connection || !connection.linked) return { tone: 'neutral', label: 'not linked', dot: false };
	return { tone: 'success', label: 'linked', dot: true };
}

export const OAUTH_ERRORS: Record<string, string> = {
	invalid_state: 'That sign-in link expired. Start the connection again.',
	forbidden: 'Your permission to manage chat connections changed before the install finished.',
	exchange_failed: 'The provider rejected the install. Try connecting again.',
	not_configured: 'This provider is not configured on the server yet.',
	secret_unavailable: 'Secret storage is not configured, so the token could not be saved.',
	denied: 'The install was cancelled.',
	error: 'The connection could not be completed.'
};

export function previewChannelName(
	pattern: string,
	ctx: { number: number; slug: string } = { number: 2481, slug: 'checkout-degraded' }
): string {
	return pattern.replaceAll('{number}', String(ctx.number)).replaceAll('{slug}', ctx.slug);
}
