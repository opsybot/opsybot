<script lang="ts">
	import GitBranchIcon from '@lucide/svelte/icons/git-branch';
	import PlugIcon from '@lucide/svelte/icons/plug';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import { ws } from '$lib/navigation';

	let { children } = $props();

	const showTabs = $derived(
		page.url.pathname === ws('/alert-sources') || page.url.pathname === ws('/alert-sources/routing')
	);
	const current = $derived(page.url.pathname === ws('/alert-sources/routing') ? 'routing' : 'sources');

	const TABS = [
		{ id: 'sources', label: 'Sources', icon: PlugIcon, href: '/alert-sources' },
		{ id: 'routing', label: 'Routing rules', icon: GitBranchIcon, href: '/alert-sources/routing' }
	] as const;
</script>

<Page title="Alert sources" subtitle="Sources in, routing rules out">
	{#if showTabs}
		<nav aria-label="Alert sources views" class="flex w-full items-center gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]">
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
	{:else}
		<div class="mt-[18px]">
			{@render children()}
		</div>
	{/if}
</Page>
