<script lang="ts">
	import EyeIcon from '@lucide/svelte/icons/eye';
	import PlugIcon from '@lucide/svelte/icons/plug';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';

	let { children } = $props();

	const current = $derived(
		page.url.pathname === '/ai/prompts' ? 'prompts' : page.url.pathname === '/ai/surfaces' ? 'surfaces' : 'models'
	);

	const TABS = [
		{ id: 'models', label: 'Model connections', icon: PlugIcon, href: '/ai' },
		{ id: 'prompts', label: 'Prompt transparency', icon: EyeIcon, href: '/ai/prompts' },
		{ id: 'surfaces', label: 'AI surfaces', icon: SparklesIcon, href: '/ai/surfaces' }
	] as const;
</script>

<Page title="AI settings" subtitle="Your models, your data, auditable prompts">
	<nav aria-label="AI settings views" class="flex w-full items-center gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]">
		{#each TABS as tab (tab.id)}
			{@const active = tab.id === current}
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
	<div class="mt-[18px]">
		{@render children()}
	</div>
</Page>
