<script lang="ts">
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import { ws } from '$lib/navigation';

	let { current }: { current: 'schedules' | 'mine' } = $props();

	const TABS = [
		{ id: 'schedules', label: 'Schedules', icon: CalendarClockIcon, href: '/on-call' },
		{ id: 'mine', label: 'My on-call', icon: UserIcon, href: '/on-call/mine' }
	];
</script>

<nav aria-label="On-call views" class="flex w-full items-center gap-1 border-b">
	{#each TABS as tab (tab.id)}
		{@const active = tab.id === current}
		<a
			href={ws(tab.href)}
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
