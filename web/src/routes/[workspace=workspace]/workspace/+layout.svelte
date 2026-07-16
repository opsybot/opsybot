<script lang="ts">
	import BracesIcon from '@lucide/svelte/icons/braces';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import UsersIcon from '@lucide/svelte/icons/users';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import { ws } from '$lib/navigation';

	let { children } = $props();

	const TABS = [
		{ id: 'members', label: 'Members', icon: UsersIcon, href: '/workspace' },
		{ id: 'teams', label: 'Teams', icon: UsersIcon, href: '/workspace/teams' },
		{ id: 'keys', label: 'API keys', icon: KeyRoundIcon, href: '/workspace/keys' },
		{ id: 'audit', label: 'Audit log', icon: HistoryIcon, href: '/workspace/audit' },
		{ id: 'settings', label: 'Settings', icon: SettingsIcon, href: '/workspace/settings' },
		{ id: 'config', label: 'Config as code', icon: BracesIcon, href: '/workspace/config' }
	] as const;

	const path = $derived(page.url.pathname);
	const isActive = (href: string) =>
		href === '/workspace'
			? path === ws('/workspace')
			: path === ws(href) || path.startsWith(`${ws(href)}/`);

	const showTabs = $derived(!path.startsWith(`${ws('/workspace/teams')}/`));
</script>

<Page title="Workspace admin" subtitle="Members, access, and the paper trail">
	{#if showTabs}
		<nav aria-label="Workspace admin views" class="flex w-full items-center gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]">
			{#each TABS as tab (tab.id)}
				{@const active = isActive(tab.href)}
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
