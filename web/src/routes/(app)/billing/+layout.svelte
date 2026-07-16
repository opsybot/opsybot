<script lang="ts">
	import CreditCardIcon from '@lucide/svelte/icons/credit-card';
	import KeyRoundIcon from '@lucide/svelte/icons/key-round';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import RadioIcon from '@lucide/svelte/icons/radio';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';

	let { data, children } = $props();

	const CLOUD_TABS = [
		{ id: 'plans', label: 'Plans', icon: LayoutDashboardIcon, href: '/billing/plans' },
		{ id: 'account', label: 'Billing', icon: CreditCardIcon, href: '/billing/account' }
	];
	const SELF_TABS = [
		{ id: 'license', label: 'License', icon: KeyRoundIcon, href: '/billing/license' },
		{ id: 'delivery', label: 'Delivery bridge', icon: RadioIcon, href: '/billing/delivery' }
	];

	const cloud = $derived(data.deployment === 'cloud');
	const tabs = $derived(cloud ? CLOUD_TABS : SELF_TABS);
	const path = $derived(page.url.pathname);
	const isActive = (href: string) => path === href || path.startsWith(`${href}/`);
	const showTabs = $derived(path !== '/billing/cancel');
</script>

<Page title={cloud ? 'Billing and plans' : 'License'} subtitle="Flat pricing, unlimited responders">
	{#if showTabs}
		<nav aria-label="Billing views" class="flex w-full items-center gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]">
			{#each tabs as tab (tab.id)}
				{@const active = isActive(tab.href)}
				<a
					href={tab.href}
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
	{/if}
	<div class="mt-[18px]">
		{@render children()}
	</div>
</Page>
