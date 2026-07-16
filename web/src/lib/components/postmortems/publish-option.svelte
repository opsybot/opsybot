<script lang="ts">
	import type { Snippet } from 'svelte';
	import { tick, untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import { Switch } from '$lib/components/ui/switch';

	let {
		option,
		on,
		label,
		disabled = false,
		onchanged,
		children
	}: {
		option: 'announce' | 'publicLink';
		on: boolean;
		label: string;
		disabled?: boolean;
		onchanged?: (on: boolean) => void;
		children: Snippet;
	} = $props();

	let checked = $state(untrack(() => on));
	$effect(() => {
		checked = on;
	});

	let form = $state<HTMLFormElement | null>(null);
	let wanted = $state(untrack(() => on));
</script>

<form method="POST" action="?/options" bind:this={form} use:enhance class="flex items-center gap-2.5">
	<input type="hidden" name="option" value={option} />
	<input type="hidden" name="on" value={String(wanted)} />

	<Switch
		bind:checked
		{disabled}
		aria-label={label}
		onCheckedChange={async (next) => {
			// Await tick so the hidden input flushes before requestSubmit reads it
			wanted = next;
			await tick();
			form?.requestSubmit();
			onchanged?.(next);
		}}
	/>

	<span class="text-muted-foreground text-[13px]">{@render children()}</span>
</form>
