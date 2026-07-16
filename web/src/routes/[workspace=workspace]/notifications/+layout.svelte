<script lang="ts">
	import BellIcon from '@lucide/svelte/icons/bell';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import { ws } from '$lib/navigation';

	let { children } = $props();

	const current = $derived(page.url.pathname === ws('/notifications/rules') ? 'rules' : 'channels');

	const TABS = [
		{ id: 'channels', label: 'My channels', icon: BellIcon, href: '/notifications' },
		{ id: 'rules', label: 'My rules', icon: ListChecksIcon, href: '/notifications/rules' }
	] as const;
</script>

<Page title="My notifications" subtitle="How Opsybot reaches you">
	<nav aria-label="Notification views" class="flex w-full items-center gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]">
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
