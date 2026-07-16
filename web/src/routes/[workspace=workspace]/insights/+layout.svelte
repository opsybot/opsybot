<script lang="ts">
	import BellIcon from '@lucide/svelte/icons/bell';
	import BookOpenIcon from '@lucide/svelte/icons/book-open';
	import ChartLineIcon from '@lucide/svelte/icons/chart-line';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import ScaleIcon from '@lucide/svelte/icons/scale';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import { ws } from '$lib/navigation';

	let { children } = $props();

	const TABS = [
		{ id: 'overview', label: 'Overview', icon: ChartLineIcon, href: '/insights' },
		{ id: 'alerts', label: 'Alert analytics', icon: BellIcon, href: '/insights/alerts' },
		{ id: 'load', label: 'On-call load', icon: ScaleIcon, href: '/insights/load' },
		{ id: 'followups', label: 'Follow-up completion', icon: ListChecksIcon, href: '/insights/followups' },
		{ id: 'definitions', label: 'Definitions', icon: BookOpenIcon, href: '/insights/definitions' }
	];

	const current = $derived(
		page.url.pathname === ws('/insights') ? 'overview' : page.url.pathname.split('/')[3]
	);
</script>

<Page title="Insights" subtitle="Numbers over adjectives">
	<nav aria-label="Insights views" class="flex w-full items-center gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]">
		{#each TABS as tab (tab.id)}
			{@const active = tab.id === current}
			<a
				href={ws(tab.href)}
				aria-current={active ? 'page' : undefined}
				class="focus-visible:ring-ring/50 inline-flex items-center gap-[7px] border-b-2 px-3 py-2.5 text-sm font-medium whitespace-nowrap transition-colors duration-[120ms] ease-out outline-none focus-visible:ring-2 [&_svg]:size-4 [&_svg]:shrink-0
				{active
					? 'border-primary text-foreground font-semibold'
					: 'text-subtle-foreground hover:text-muted-foreground border-transparent'}"
			>
				<tab.icon />
				{tab.label}
			</a>
		{/each}
	</nav>

	<div class="mt-[18px]">
		{@render children()}
	</div>
</Page>
