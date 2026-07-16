<script lang="ts">
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import GlobeIcon from '@lucide/svelte/icons/globe';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';

	let { base, current }: { base: string; current: string } = $props();

	const TABS = [
		{ value: '', label: 'Overview', icon: LayoutDashboardIcon },
		{ value: '/timeline', label: 'Timeline', icon: ActivityIcon },
		{ value: '/follow-ups', label: 'Follow-ups', icon: ListChecksIcon },
		{ value: '/status-page', label: 'Status page', icon: GlobeIcon },
		{ value: '/postmortem', label: 'Postmortem', icon: FileTextIcon }
	];
</script>

<nav aria-label="Incident views" class="flex w-full items-center gap-1 border-b">
	{#each TABS as tab (tab.value)}
		{@const active = tab.value === current}
		<a
			href={base + tab.value}
			aria-current={active ? 'page' : undefined}
			class="focus-visible:ring-ring/50 -mb-px inline-flex items-center gap-[7px] border-b-2 px-3 py-2.5 text-sm font-medium whitespace-nowrap transition-colors duration-[120ms] ease-out outline-none focus-visible:ring-2 [&_svg]:size-4 [&_svg]:shrink-0
			{active
				? 'border-primary text-foreground font-semibold'
				: 'text-subtle-foreground hover:text-muted-foreground border-transparent'}"
		>
			<tab.icon />
			{tab.label}
		</a>
	{/each}
</nav>
