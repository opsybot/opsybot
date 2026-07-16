<script lang="ts">
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import SirenIcon from '@lucide/svelte/icons/siren';

	let { current }: { current: 'list' | 'follow-ups' } = $props();

	const TABS = [
		{ id: 'list', label: 'Incidents', icon: SirenIcon, href: '/incidents' },
		{ id: 'follow-ups', label: 'Follow-ups', icon: ListChecksIcon, href: '/incidents/follow-ups' }
	];
</script>

<nav aria-label="Incidents views" class="flex w-full items-center gap-1 border-b">
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
