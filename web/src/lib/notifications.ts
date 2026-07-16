import type { Tone } from '$lib/dashboard';

export type ChannelType = 'slack' | 'teams' | 'discord' | 'telegram' | 'ntfy' | 'email' | 'webhook';

export type ConnectKind = 'oauth' | 'telegram' | 'ntfy' | 'email' | 'webhook';

export type ChannelMeta = {
	id: ChannelType;
	label: string;
	icon: string;
	desc: string;
	connect: ConnectKind;
};

export const CHANNEL_TYPES: ChannelMeta[] = [
	{ id: 'slack', label: 'Slack DM', icon: 'message-square', desc: 'DM from the Opsybot app', connect: 'oauth' },
	{ id: 'teams', label: 'Microsoft Teams', icon: 'message-square', desc: 'Chat from the Opsybot bot', connect: 'oauth' },
	{ id: 'discord', label: 'Discord', icon: 'message-square', desc: 'DM from the Opsybot bot', connect: 'oauth' },
	{ id: 'telegram', label: 'Telegram', icon: 'send', desc: 'Message from @opsybot_bot', connect: 'telegram' },
	{ id: 'ntfy', label: 'ntfy', icon: 'bell', desc: 'Push to an ntfy topic', connect: 'ntfy' },
	{ id: 'email', label: 'Email', icon: 'mail', desc: 'Plain email, one per page', connect: 'email' },
	{ id: 'webhook', label: 'Webhook', icon: 'webhook', desc: 'POST to your own endpoint', connect: 'webhook' }
];

export type Channel = { id: string; type: ChannelType; detail: string; verified: boolean };

export type RuleStep = { id: string; channel: ChannelType; delay: string };

export type QuietHours = {
	enabled: boolean;
	start: string;
	end: string;
	timezone: string;
	days: string[];
};

export const DELAY_OPTIONS = [
	{ value: '0', label: 'immediately' },
	{ value: '1', label: 'after 1 min' },
	{ value: '2', label: 'after 2 min' },
	{ value: '5', label: 'after 5 min' },
	{ value: '10', label: 'after 10 min' }
];

export const HOUR_OPTIONS = Array.from({ length: 24 }, (_, hour) => {
	const value = `${String(hour).padStart(2, '0')}:00`;
	return { value, label: value };
});

export const TIMEZONE_OPTIONS = ['Europe/Berlin (device)', 'UTC'];

export const DAYS: { value: string; full: string }[] = [
	{ value: 'Mon', full: 'Monday' },
	{ value: 'Tue', full: 'Tuesday' },
	{ value: 'Wed', full: 'Wednesday' },
	{ value: 'Thu', full: 'Thursday' },
	{ value: 'Fri', full: 'Friday' },
	{ value: 'Sat', full: 'Saturday' },
	{ value: 'Sun', full: 'Sunday' }
];

export const DEFAULT_QUIET_HOURS: QuietHours = {
	enabled: true,
	start: '22:00',
	end: '07:00',
	timezone: 'Europe/Berlin (device)',
	days: DAYS.map((day) => day.value)
};

const CHANNEL_IDS = new Set<string>(CHANNEL_TYPES.map((channel) => channel.id));
const DELAY_VALUES = new Set(DELAY_OPTIONS.map((delay) => delay.value));
const HOUR_VALUES = new Set(HOUR_OPTIONS.map((hour) => hour.value));
const DAY_VALUES = new Set(DAYS.map((day) => day.value));
export const MAX_STEPS = 12;
const DEFAULT_LATER_DELAY = '5';

export function isChannelType(value: string): value is ChannelType {
	return CHANNEL_IDS.has(value);
}

export function channelMeta(type: ChannelType): ChannelMeta {
	return CHANNEL_TYPES.find((channel) => channel.id === type) ?? CHANNEL_TYPES[0];
}

export function channelLabel(type: string): string {
	return CHANNEL_TYPES.find((channel) => channel.id === type)?.label ?? type;
}

export function verifyBadge(channel: Pick<Channel, 'verified'>): { tone: Tone; label: string } {
	return channel.verified
		? { tone: 'success', label: 'verified' }
		: { tone: 'warning', label: 'unverified' };
}

export function previewSentence(steps: RuleStep[]): string {
	if (!steps.length) return 'You would get nothing — add at least one step.';
	const parts = steps.map((step, index) => {
		const when =
			index === 0 ? 'immediately' : step.delay === '0' ? 'at the same time' : `after ${step.delay} min`;
		return `${channelLabel(step.channel)} ${when}`;
	});
	const joined =
		parts.length === 1 ? parts[0] : `${parts.slice(0, -1).join(', ')}, then ${parts[parts.length - 1]}`;
	return `You'll get ${joined}.`;
}

let idSeq = 0;
export function uid(): string {
	idSeq += 1;
	return `s${idSeq.toString(36)}-${Math.random().toString(36).slice(2, 7)}`;
}

// First step delay is '0', later steps positive; re-enforced after reorders
export function normalizeSteps(steps: RuleStep[]): RuleStep[] {
	return steps.map((step, index) => {
		const delay = index === 0 ? '0' : step.delay === '0' ? DEFAULT_LATER_DELAY : step.delay;
		return delay === step.delay ? step : { ...step, delay };
	});
}

function parseSteps(input: unknown): RuleStep[] {
	if (!Array.isArray(input)) return [];
	const out: RuleStep[] = [];
	for (const raw of input.slice(0, MAX_STEPS)) {
		if (!raw || typeof raw !== 'object') continue;
		const step = raw as Record<string, unknown>;
		const channel = typeof step.channel === 'string' && isChannelType(step.channel) ? step.channel : 'email';
		const delay = typeof step.delay === 'string' && DELAY_VALUES.has(step.delay) ? step.delay : '0';
		out.push({ id: uid(), channel, delay });
	}
	return normalizeSteps(out);
}

function parseQuietHours(input: unknown): QuietHours {
	const raw = input && typeof input === 'object' ? (input as Record<string, unknown>) : {};
	const submitted = Array.isArray(raw.days)
		? raw.days.filter((day): day is string => typeof day === 'string' && DAY_VALUES.has(day))
		: [];
	return {
		enabled: raw.enabled === true,
		start: typeof raw.start === 'string' && HOUR_VALUES.has(raw.start) ? raw.start : DEFAULT_QUIET_HOURS.start,
		end: typeof raw.end === 'string' && HOUR_VALUES.has(raw.end) ? raw.end : DEFAULT_QUIET_HOURS.end,
		timezone:
			typeof raw.timezone === 'string' && TIMEZONE_OPTIONS.includes(raw.timezone)
				? raw.timezone
				: TIMEZONE_OPTIONS[0],
		days: DAYS.map((day) => day.value).filter((day) => submitted.includes(day))
	};
}

// Sanitizes client-submitted rules; the builder UI is not a trust boundary
export function parseRules(
	raw: string
): { high: RuleStep[]; low: RuleStep[]; quietHours: QuietHours } | { error: string } {
	let data: unknown;
	try {
		data = JSON.parse(raw);
	} catch {
		return { error: 'Could not read the rules.' };
	}
	if (!data || typeof data !== 'object') return { error: 'Could not read the rules.' };
	const obj = data as Record<string, unknown>;
	return { high: parseSteps(obj.high), low: parseSteps(obj.low), quietHours: parseQuietHours(obj.quietHours) };
}
