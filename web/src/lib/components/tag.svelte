<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		selected = false,
		dot,
		href,
		onclick,
		onremove,
		class: className,
		children
	}: {
		selected?: boolean;
		dot?: string;
		href?: string;
		onclick?: () => void;
		onremove?: () => void;
		class?: string;
		children: Snippet;
	} = $props();

	const classes = $derived(
		cn(
			'inline-flex h-6 items-center gap-1.5 rounded-md border pl-2.5 text-xs leading-none font-medium whitespace-nowrap',
			onremove ? 'pr-1.5' : 'pr-2.5',
			selected
				? 'bg-brand-wash border-brand-edge text-brand-foreground'
				: 'bg-popover border-input text-muted-foreground',
			href && 'hover:border-brand-edge hover:text-brand-foreground transition-colors',
			className
		)
	);
</script>

{#snippet body()}
	{#if dot}
		<span class="size-1.5 shrink-0 rounded-full" style="background: {dot}"></span>
	{/if}
	{@render children()}
	{#if onremove}
		<button
			type="button"
			aria-label="Remove"
			class="ml-px inline-flex size-4 items-center justify-center rounded-[3px] text-[13px] leading-none opacity-60 hover:bg-white/8 hover:opacity-100"
			onclick={(event) => {
				event.stopPropagation();
				onremove();
			}}
		>
			×
		</button>
	{/if}
{/snippet}

{#if href}
	<a {href} class={classes}>{@render body()}</a>
{:else if onclick}
	<button type="button" {onclick} class={classes} aria-pressed={selected}>
		{@render body()}
	</button>
{:else}
	<span class={classes}>{@render body()}</span>
{/if}
