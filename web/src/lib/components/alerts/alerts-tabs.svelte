<script lang="ts">
	import BellIcon from '@lucide/svelte/icons/bell';
	import BellOffIcon from '@lucide/svelte/icons/bell-off';
	import HeartPulseIcon from '@lucide/svelte/icons/heart-pulse';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';

	let { current }: { current: 'list' | 'silences' | 'failures' | 'heartbeats' } = $props();

	const TABS = [
		{ id: 'list', label: 'Alerts', icon: BellIcon, href: '/alerts' },
		{ id: 'silences', label: 'Silences', icon: BellOffIcon, href: '/alerts/silences' },
		{
			id: 'failures',
			label: 'Ingestion failures',
			icon: TriangleAlertIcon,
			href: '/alerts/failures'
		},
		{ id: 'heartbeats', label: 'Heartbeats', icon: HeartPulseIcon, href: '/alerts/heartbeats' }
	];
</script>

<nav aria-label="Alerts views" class="flex w-full items-center gap-1 border-b">
	{#each TABS as tab (tab.id)}
		{@const active = tab.id === current}
		<a
			href={tab.href}
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
