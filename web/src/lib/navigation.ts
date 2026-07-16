import type { Component } from 'svelte';
import type { LucideProps } from '@lucide/svelte';
import ArrowUpRightIcon from '@lucide/svelte/icons/arrow-up-right';
import BellIcon from '@lucide/svelte/icons/bell';
import BellRingIcon from '@lucide/svelte/icons/bell-ring';
import BoxesIcon from '@lucide/svelte/icons/boxes';
import Building2Icon from '@lucide/svelte/icons/building-2';
import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
import ChartLineIcon from '@lucide/svelte/icons/chart-line';
import CreditCardIcon from '@lucide/svelte/icons/credit-card';
import FileTextIcon from '@lucide/svelte/icons/file-text';
import GlobeIcon from '@lucide/svelte/icons/globe';
import HouseIcon from '@lucide/svelte/icons/house';
import ImportIcon from '@lucide/svelte/icons/import';
import MessageSquareIcon from '@lucide/svelte/icons/message-square';
import PlugIcon from '@lucide/svelte/icons/plug';
import ServerIcon from '@lucide/svelte/icons/server';
import SettingsIcon from '@lucide/svelte/icons/settings';
import SirenIcon from '@lucide/svelte/icons/siren';
import SparklesIcon from '@lucide/svelte/icons/sparkles';
import WorkflowIcon from '@lucide/svelte/icons/workflow';
import { page } from '$app/state';

export type NavItem = {
	title: string;
	href: string;
	icon: Component<LucideProps>;
};

export type NavSection = {
	label?: string;
	items: NavItem[];
};

export const navigation: NavSection[] = [
	{
		items: [
			{ title: 'Home', href: '/', icon: HouseIcon },
			{ title: 'Incidents', href: '/incidents', icon: SirenIcon },
			{ title: 'Alerts', href: '/alerts', icon: BellIcon },
			{ title: 'On-call', href: '/on-call', icon: CalendarClockIcon },
			{ title: 'Postmortems', href: '/postmortems', icon: FileTextIcon }
		]
	},
	{
		label: 'Foundation',
		items: [
			{ title: 'Catalog', href: '/catalog', icon: BoxesIcon },
			{ title: 'Status pages', href: '/status-pages', icon: GlobeIcon },
			{ title: 'Workflows', href: '/workflows', icon: WorkflowIcon },
			{ title: 'Insights', href: '/insights', icon: ChartLineIcon }
		]
	},
	{
		label: 'Configure',
		items: [
			{ title: 'Alert sources', href: '/alert-sources', icon: PlugIcon },
			{ title: 'Escalation policies', href: '/escalation-policies', icon: ArrowUpRightIcon },
			{ title: 'Chat connections', href: '/chat', icon: MessageSquareIcon },
			{ title: 'My notifications', href: '/notifications', icon: BellRingIcon },
			{ title: 'AI settings', href: '/ai', icon: SparklesIcon },
			{ title: 'Workspace admin', href: '/workspace', icon: SettingsIcon },
			{ title: 'Billing and plans', href: '/billing', icon: CreditCardIcon },
			{ title: 'Instance operations', href: '/operations', icon: ServerIcon },
			{ title: 'Import from Opsgenie', href: '/import', icon: ImportIcon },
			{ title: 'Enterprise', href: '/enterprise', icon: Building2Icon }
		]
	}
];

export function ws(path = ''): string {
	return `/${page.params.workspace}${path === '/' ? '' : path}`;
}

export function workspacePath(pathname: string, workspace: string): string {
	return pathname.slice(`/${workspace}`.length) || '/';
}

export function isCurrentSection(pathname: string, href: string): boolean {
	if (href === '/') return pathname === '/';
	return pathname === href || pathname.startsWith(`${href}/`);
}
