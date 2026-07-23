import type { Component } from 'svelte';
import type { LucideProps } from '@lucide/svelte';
import BellIcon from '@lucide/svelte/icons/bell';
import BoxesIcon from '@lucide/svelte/icons/boxes';
import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
import MessageSquareIcon from '@lucide/svelte/icons/message-square';
import PlugIcon from '@lucide/svelte/icons/plug';

export type OnboardingStepId = 'schedule' | 'notify' | 'source' | 'chat' | 'service';

export type OnboardingStep = {
	id: OnboardingStepId;
	icon: Component<LucideProps>;
	title: string;
	description: string;
	action: string;
	href: string;
};

export const ONBOARDING_STEPS: OnboardingStep[] = [
	{
		id: 'schedule',
		icon: CalendarClockIcon,
		title: 'Create your first schedule',
		description: "Who's on call, and when. Rotations and overrides come free.",
		action: 'Create schedule',
		href: '/on-call'
	},
	{
		id: 'notify',
		icon: BellIcon,
		title: 'Connect a notification channel',
		description: 'Add your phone or push, then send yourself one test page.',
		action: 'Send test page',
		href: '/notifications'
	},
	{
		id: 'source',
		icon: PlugIcon,
		title: 'Connect an alert source',
		description: 'Point Prometheus, Datadog, Grafana, or plain email at Opsybot.',
		action: 'Connect source',
		href: '/alert-sources'
	},
	{
		id: 'chat',
		icon: MessageSquareIcon,
		title: 'Connect chat',
		description: 'Incidents run in Slack, Microsoft Teams, or Discord.',
		action: 'Connect chat',
		href: '/integrations'
	},
	{
		id: 'service',
		icon: BoxesIcon,
		title: 'Create a service',
		description: 'Services tie alerts, owners, and status pages together.',
		action: 'Create service',
		href: '/catalog'
	}
];
