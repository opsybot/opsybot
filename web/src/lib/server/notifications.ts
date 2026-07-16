import type { Channel, ChannelType, QuietHours, RuleStep } from '$lib/notifications';
import { DEFAULT_QUIET_HOURS, isChannelType, uid } from '$lib/notifications';
import { scenario } from './fixtures';

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

export function getRules(): { high: RuleStep[]; low: RuleStep[]; quietHours: QuietHours } {
	return { high: store.high, low: store.low, quietHours: store.quietHours };
}

export function saveRules(high: RuleStep[], low: RuleStep[], quietHours: QuietHours): void {
	store.high = high;
	store.low = low;
	store.quietHours = quietHours;
}
