<script lang="ts">
	import type { Snippet } from 'svelte';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';

	let { children }: { children: Snippet } = $props();
</script>

<div
	class="border-border-strong bg-inset text-subtle-foreground flex items-center gap-2.5 rounded-xl border border-dashed px-4 py-3 text-[12.5px]"
>
	<SparklesIcon class="size-3.5 shrink-0" />
	<span>{@render children()}</span>
</div>
