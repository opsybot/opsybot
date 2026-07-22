import type { Cookies } from '@sveltejs/kit';
import type { components } from '$lib/api/schema';
import type { Channel, ChannelType, QuietHours, RuleStep } from '$lib/notifications';
import { DAYS, DELAY_OPTIONS, DEFAULT_QUIET_HOURS, HOUR_OPTIONS, TIMEZONE_OPTIONS, isChannelType, uid } from '$lib/notifications';
import { apiClient } from './api';
import { scenario } from './fixtures';

type Schemas = components['schemas'];

const ISO_WEEKDAY: Record<string, number> = { Sun: 0, Mon: 1, Tue: 2, Wed: 3, Thu: 4, Fri: 5, Sat: 6 };
const SHORT_WEEKDAY: Record<number, string> = { 0: 'Sun', 1: 'Mon', 2: 'Tue', 3: 'Wed', 4: 'Thu', 5: 'Fri', 6: 'Sat' };

const DELAY_VALUES = DELAY_OPTIONS.map((option) => Number(option.value));

function nearestDelay(minutes: number): string {
	let best = DELAY_VALUES[0];
	for (const value of DELAY_VALUES) {
		if (Math.abs(value - minutes) < Math.abs(best - minutes)) best = value;
	}
	return String(best);
}

function minutesToHour(minutes: number): string {
	const hour = Math.min(23, Math.max(0, Math.floor(minutes / 60)));
	const candidate = `${String(hour).padStart(2, '0')}:00`;
	return HOUR_OPTIONS.some((option) => option.value === candidate) ? candidate : HOUR_OPTIONS[0].value;
}

function hourToMinutes(value: string): number {
	const [hours] = value.split(':');
	const hour = Number(hours);
	return Number.isFinite(hour) ? Math.min(23, Math.max(0, hour)) * 60 : 0;
}

function normalizeTimezone(timezone: string): string {
	return timezone.replace(/ \(device\)$/, '').trim() || 'UTC';
}

function displayTimezone(timezone: string): string {
	if (!timezone) return DEFAULT_QUIET_HOURS.timezone;
	if (TIMEZONE_OPTIONS.includes(timezone)) return timezone;
	const withSuffix = `${timezone} (device)`;
	return TIMEZONE_OPTIONS.includes(withSuffix) ? withSuffix : timezone;
}

function stepsFromApi(steps: Schemas['NotificationStep'][]): RuleStep[] {
	return steps.map((step) => ({ id: uid(), channel: step.channelType as ChannelType, delay: nearestDelay(step.delayMinutes) }));
}

function stepsToApi(steps: RuleStep[]): Schemas['NotificationStep'][] {
	return steps.map((step) => ({ channelType: step.channel, delayMinutes: Number(step.delay) || 0 }));
}

function quietFromApi(quiet: Schemas['NotificationQuietHours']): QuietHours {
	const days = quiet.days.map((day) => SHORT_WEEKDAY[day]).filter((day): day is string => Boolean(day));
	return {
		enabled: quiet.enabled,
		start: minutesToHour(quiet.startMinute),
		end: minutesToHour(quiet.endMinute),
		timezone: displayTimezone(quiet.timezone),
		days: DAYS.map((day) => day.value).filter((day) => days.includes(day))
	};
}

function quietToApi(quiet: QuietHours): Schemas['NotificationQuietHours'] {
	return {
		enabled: quiet.enabled,
		days: quiet.days.map((day) => ISO_WEEKDAY[day]).filter((day) => day !== undefined),
		startMinute: hourToMinutes(quiet.start),
		endMinute: hourToMinutes(quiet.end),
		timezone: normalizeTimezone(quiet.timezone)
	};
}

function channelFromApi(dto: Schemas['Channel']): Channel {
	return { id: dto.id, type: dto.type as ChannelType, detail: dto.detail, verified: dto.verified };
}

const CONNECT_DETAIL: Record<ChannelType, string> = {
	slack: 'Acme Corp · @maya',
	teams: 'Acme Corp · Maya Chen',
	discord: 'maya#4821',
	telegram: '@mayachen',
	ntfy: 'ntfy.sh/maya-pages-x7k2',
	email: 'maya@acme.dev',
	webhook: 'hooks.example.dev/page'
};

function step(channel: ChannelType, delay: string): RuleStep {
	return { id: uid(), channel, delay };
}

function seed() {
	const channels: Channel[] = [
		{ id: 'ntfy', type: 'ntfy', detail: CONNECT_DETAIL.ntfy, verified: true },
		{ id: 'telegram', type: 'telegram', detail: CONNECT_DETAIL.telegram, verified: true },
		{ id: 'email', type: 'email', detail: CONNECT_DETAIL.email, verified: true },
		{ id: 'slack', type: 'slack', detail: CONNECT_DETAIL.slack, verified: false }
	];
	const high: RuleStep[] = [step('ntfy', '0'), step('telegram', '2'), step('email', '5')];
	const low: RuleStep[] = [step('email', '0')];
	const quietHours: QuietHours = { ...DEFAULT_QUIET_HOURS };
	return { channels, high, low, quietHours };
}

const store = seed();

const state = scenario();
if (state === 'empty') {
	store.channels = [];
	store.high = [];
	store.low = [];
	store.quietHours = { ...DEFAULT_QUIET_HOURS, enabled: false };
}
if (state === 'degraded') {
	const ntfy = store.channels.find((channel) => channel.id === 'ntfy');
	if (ntfy) ntfy.verified = false;
}

export function listChannels(): Channel[] {
	return store.channels;
}

export function connectChannel(type: string): boolean {
	if (!isChannelType(type)) return false;
	const existing = store.channels.find((channel) => channel.type === type);
	if (existing) {
		existing.verified = true;
		return true;
	}
	store.channels.push({ id: type, type, detail: CONNECT_DETAIL[type], verified: true });
	return true;
}

export function removeChannel(id: string): boolean {
	const index = store.channels.findIndex((channel) => channel.id === id);
	if (index < 0) return false;
	store.channels.splice(index, 1);
	return true;
}

export async function getRules(
	cookies: Cookies,
	workspace: string
): Promise<{ high: RuleStep[]; low: RuleStep[]; quietHours: QuietHours; channels: Channel[] }> {
	const { data } = await apiClient(cookies).GET('/workspaces/{workspaceId}/notification-rules', {
		params: { path: { workspaceId: workspace } }
	});
	if (!data) {
		return { high: [], low: [], quietHours: { ...DEFAULT_QUIET_HOURS }, channels: [] };
	}
	return {
		high: stepsFromApi(data.rules.high),
		low: stepsFromApi(data.rules.low),
		quietHours: quietFromApi(data.rules.quietHours),
		channels: data.channels.map(channelFromApi)
	};
}

export async function saveRules(
	cookies: Cookies,
	workspace: string,
	high: RuleStep[],
	low: RuleStep[],
	quietHours: QuietHours
): Promise<{ error?: string }> {
	const { error } = await apiClient(cookies).PUT('/workspaces/{workspaceId}/notification-rules', {
		params: { path: { workspaceId: workspace } },
		body: { high: stepsToApi(high), low: stepsToApi(low), quietHours: quietToApi(quietHours) }
	});
	return error ? { error: error.detail ?? 'Could not save your notification rules.' } : {};
}
