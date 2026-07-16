<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		title,
		aside,
		class: className,
		children
	}: {
		title?: string;
		aside?: Snippet;
		class?: string;
		children: Snippet;
	} = $props();
</script>

<section class={cn('bg-card overflow-hidden rounded-xl border', className)}>
	{#if title}
		<header class="flex items-center gap-2 border-b px-4 py-3">
			<span class="text-[13.5px] font-semibold">{title}</span>
			{@render aside?.()}
		</header>
	{/if}
	{@render children()}
</section>
