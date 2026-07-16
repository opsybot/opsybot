<script lang="ts">
	import GlobeIcon from '@lucide/svelte/icons/globe';
	import MegaphoneIcon from '@lucide/svelte/icons/megaphone';
	import UsersIcon from '@lucide/svelte/icons/users';
	import WrenchIcon from '@lucide/svelte/icons/wrench';

	let { current }: { current: 'pages' | 'publish' | 'maintenance' | 'subscribers' } = $props();

	const TABS = [
		{ id: 'pages', label: 'Pages', icon: GlobeIcon, href: '/status-pages' },
		{ id: 'publish', label: 'Publish incident', icon: MegaphoneIcon, href: '/status-pages/publish' },
		{ id: 'maintenance', label: 'Maintenance', icon: WrenchIcon, href: '/status-pages/maintenance' },
		{ id: 'subscribers', label: 'Subscribers', icon: UsersIcon, href: '/status-pages/subscribers' }
	];
</script>

<nav aria-label="Status pages views" class="flex w-full items-center gap-1 border-b">
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
