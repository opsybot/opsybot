<script lang="ts">
	import type { Snippet } from 'svelte';
	import BellIcon from '@lucide/svelte/icons/bell';
	import SirenIcon from '@lucide/svelte/icons/siren';
	import { Button } from '$lib/components/ui/button';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import CommandTrigger from './command-trigger.svelte';
	import { useAppShell } from './context.svelte';
	import ThemeToggle from './theme-toggle.svelte';

	let {
		title,
		subtitle,
		actions,
		children
	}: {
		title: string;
		subtitle?: string;
		actions?: Snippet;
		children: Snippet;
	} = $props();

	const shell = useAppShell();
	const workspace = $derived(shell.session.activeWorkspace);
	const description = $derived(subtitle ?? `${workspace.name} · ${workspace.environment}`);
</script>

<svelte:head>
	<title>{title} · Opsybot</title>
</svelte:head>

<header class="flex h-14 shrink-0 items-center gap-[14px] border-b px-[22px]">
	<Sidebar.Trigger class="-ms-1 md:hidden" />

	<div class="flex min-w-0 items-baseline gap-2.5">
		<h1 class="truncate text-[17px] font-semibold tracking-[-0.01em]">{title}</h1>
		<span class="text-subtle-foreground hidden truncate text-[12.5px] sm:inline">
			{description}
		</span>
	</div>

	<div class="flex-1"></div>

	<div class="hidden sm:block">
		<CommandTrigger />
	</div>

	{@render actions?.()}

	<Button variant="destructive" size="sm" href="/incidents?declare">
		<SirenIcon data-icon="inline-start" />
		Declare
	</Button>

	<div class="bg-border h-6 w-px shrink-0"></div>

	<ThemeToggle />

	<Button variant="ghost" size="icon" aria-label="Notifications">
		<BellIcon />
	</Button>
</header>

<div class="min-h-0 flex-1 overflow-auto">
	<div
		class="motion-safe:animate-in motion-safe:fade-in flex max-w-[1160px] flex-col gap-4 p-[22px]"
	>
		{@render children()}
	</div>
</div>
