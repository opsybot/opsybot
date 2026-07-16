<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Badge, type BadgeTone } from '$lib/components/ui/badge';
	import { cn } from '$lib/utils';

	let {
		title,
		count,
		countTone = 'neutral',
		accent,
		live = false,
		aside,
		children
	}: {
		title: string;
		count?: number;
		countTone?: BadgeTone;
		accent?: string;
		live?: boolean;
		aside?: Snippet;
		children: Snippet;
	} = $props();
</script>

<section
	class={cn('bg-card overflow-hidden rounded-xl border', accent && 'border-l-[3px]')}
	style={accent ? `border-left-color: ${accent}` : undefined}
>
	<header class="flex items-center gap-2.5 border-b px-4 py-3">
		{#if live}
			<span
				class="bg-critical motion-safe:animate-pulse-critical size-[7px] shrink-0 rounded-full"
				aria-hidden="true"
			></span>
		{/if}
		<span class="text-sm font-semibold">{title}</span>
		{#if count != null}
			<Badge tone={countTone} size="sm">{count}</Badge>
		{/if}
		<div class="flex-1"></div>
		{@render aside?.()}
	</header>
	{@render children()}
</section>
