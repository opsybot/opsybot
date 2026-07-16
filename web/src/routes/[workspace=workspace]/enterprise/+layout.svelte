<script lang="ts">
	import type { Component } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import LockIcon from '@lucide/svelte/icons/lock';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import UsersIcon from '@lucide/svelte/icons/users';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import { ENT_TABS } from '$lib/enterprise';
	import { ws } from '$lib/navigation';

	let { data, children } = $props();

	const ICON: Record<string, Component<LucideProps>> = {
		users: UsersIcon,
		'shield-check': ShieldCheckIcon,
		history: HistoryIcon,
		lock: LockIcon
	};
	const TABS = ENT_TABS.map((tab) => ({ ...tab, href: `/enterprise/${tab.id}`, Icon: ICON[tab.icon] }));

	const path = $derived(page.url.pathname);
	const isActive = (href: string) => path === href || path.startsWith(`${href}/`);
</script>

<Page title="Enterprise" subtitle={data.licensed ? 'Licensed to Acme Corp' : 'Not licensed — preview'}>
	<nav aria-label="Enterprise views" class="flex w-full items-center gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]">
		{#each TABS as tab (tab.id)}
			{@const active = isActive(ws(tab.href))}
			<a
				href={ws(tab.href)}
				aria-current={active ? 'page' : undefined}
				class="focus-visible:ring-ring/50 inline-flex items-center gap-[7px] border-b-2 px-3 py-2.5 text-sm font-medium whitespace-nowrap transition-colors duration-[120ms] ease-out outline-none focus-visible:ring-2 [&_svg]:size-4 [&_svg]:shrink-0
				{active
					? 'border-primary text-foreground font-semibold'
					: 'text-subtle-foreground hover:text-muted-foreground border-transparent'}"
			>
				<tab.Icon />
				{tab.label}
			</a>
		{/each}
	</nav>
	<div class="mt-[18px]">
		{@render children()}
	</div>
</Page>
