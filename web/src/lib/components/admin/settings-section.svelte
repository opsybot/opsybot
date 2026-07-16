<script lang="ts">
	import type { Snippet } from 'svelte';

	let { title, note, children }: { title: string; note?: string; children: Snippet } = $props();
</script>

<section class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2 border-b px-4 py-3">
		<span class="text-[13.5px] font-semibold">{title}</span>
		{#if note}<span class="text-subtle-foreground text-[11.5px]">{note}</span>{/if}
	</header>
	<div class="flex flex-col gap-3.5 p-4">
		{@render children()}
	</div>
</section>
