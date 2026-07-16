import type { Component } from 'svelte';
import type { LucideProps } from '@lucide/svelte';
import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
import ClockIcon from '@lucide/svelte/icons/clock';
import FlagIcon from '@lucide/svelte/icons/flag';
import UserIcon from '@lucide/svelte/icons/user';
import UsersIcon from '@lucide/svelte/icons/users';
import WebhookIcon from '@lucide/svelte/icons/webhook';

export const ICON: Record<string, Component<LucideProps>> = {
	user: UserIcon,
	'calendar-clock': CalendarClockIcon,
	users: UsersIcon,
	webhook: WebhookIcon,
	flag: FlagIcon,
	clock: ClockIcon
};
