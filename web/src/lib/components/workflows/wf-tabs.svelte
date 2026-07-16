<script lang="ts">
	import UsersIcon from '@lucide/svelte/icons/users';
	import WorkflowIcon from '@lucide/svelte/icons/workflow';
	import { ws } from '$lib/navigation';

	let { current }: { current: 'workflows' | 'roles' } = $props();

	const TABS = [
		{ id: 'workflows', label: 'Workflows', icon: WorkflowIcon, href: '/workflows' },
		{ id: 'roles', label: 'Incident roles', icon: UsersIcon, href: '/workflows/roles' }
	] as const;
</script>

<nav aria-label="Workflows views" class="flex w-full items-center gap-1 border-b">
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
