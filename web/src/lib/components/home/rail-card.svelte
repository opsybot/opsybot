<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		footer,
		children
	}: {
		title: string;
		footer?: Snippet;
		children: Snippet;
	} = $props();
</script>

<section class="bg-card overflow-hidden rounded-xl border">
	<header class="flex items-center gap-2.5 border-b px-4 py-3">
		<span class="text-[13.5px] font-semibold">{title}</span>
	</header>
	{@render children()}
	{#if footer}
		<footer class="border-t px-4 py-2.5 text-[12.5px]">
			{@render footer()}
		</footer>
	{/if}
</section>
