<script lang="ts">
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import DatabaseIcon from '@lucide/svelte/icons/database';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import { ws } from '$lib/navigation';

	let { children } = $props();

	const TABS = [
		{ id: 'diagnostics', label: 'Diagnostics', icon: ActivityIcon, href: '/operations/diagnostics' },
		{ id: 'backup', label: 'Backup & restore', icon: DatabaseIcon, href: '/operations/backup' }
	] as const;

	const path = $derived(page.url.pathname);
	const isActive = (href: string) => path === href || path.startsWith(`${href}/`);
</script>

<Page title="Instance operations" subtitle="Self-hosted health, backups, and updates">
	<nav aria-label="Operations views" class="flex w-full items-center gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]">
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
				<tab.icon />
				{tab.label}
			</a>
		{/each}
	</nav>
	<div class="mt-[18px]">
		{@render children()}
	</div>
</Page>
