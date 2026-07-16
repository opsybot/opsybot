<script lang="ts">
	import { onMount } from 'svelte';
	import SearchIcon from '@lucide/svelte/icons/search';
	import { useAppShell } from './context.svelte';

	const shell = useAppShell();

	let shortcut = $state('Ctrl K');

	onMount(() => {
		if (navigator.userAgent.includes('Macintosh')) shortcut = '⌘K';
	});
</script>

<button
	type="button"
	data-slot="command-trigger"
	onclick={() => shell.openCommand()}
	class="bg-inset text-subtle-foreground hover:border-border-strong hover:text-muted-foreground inline-flex h-8 shrink-0 items-center gap-2 rounded-md border pr-2 pl-[11px] text-[12.5px] transition-colors"
>
	<SearchIcon class="size-[13px] shrink-0" />
	<span class="pr-[22px]">Search</span>
	<kbd
		class="bg-card border-input text-subtle-foreground rounded-[4px] border px-[5px] py-px font-mono text-[10px]"
	>
		{shortcut}
	</kbd>
</button>
