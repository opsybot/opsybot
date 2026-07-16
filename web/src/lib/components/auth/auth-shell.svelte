<script lang="ts">
	import type { Snippet } from 'svelte';
	import Wordmark from '$lib/components/layout/wordmark.svelte';

	let {
		title,
		subtitle,
		width = 400,
		footer,
		children
	}: {
		title?: string;
		subtitle?: string;
		width?: number;
		footer?: Snippet;
		children: Snippet;
	} = $props();
</script>

<svelte:head>
	<title>{title ? `${title} · Opsybot` : 'Opsybot'}</title>
</svelte:head>

<div
	class="flex min-h-svh flex-col items-center justify-center gap-7 px-5 py-12 max-[560px]:gap-[22px] max-[560px]:px-3.5 max-[560px]:py-7"
>
	<a href="/login" class="text-foreground hover:no-underline">
		<Wordmark class="text-[22px]" />
	</a>

	<div
		class="bg-card shadow-md w-full rounded-xl border p-8 max-[560px]:px-5 max-[560px]:py-[26px]"
		style="max-width: {width}px"
	>
		{#if title}
			<h1 class="mb-6 text-[26px] font-light tracking-[-0.02em]">{title}</h1>
		{/if}
		{#if subtitle}
			<p class="text-muted-foreground -mt-4 mb-6 text-sm leading-[1.55]">{subtitle}</p>
		{/if}
		{@render children()}
	</div>

	{#if footer}
		<div class="text-subtle-foreground max-w-[400px] text-center text-[13px] leading-[1.6]">
			{@render footer()}
		</div>
	{/if}
</div>
