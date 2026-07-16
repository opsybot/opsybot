<script lang="ts">
	import '../app.css';
	import { untrack } from 'svelte';
	import favicon from '$lib/assets/favicon.svg';
	import { Toaster } from '$lib/components/ui/sonner';
	import { setTheme } from '$lib/theme.svelte';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	// Seed theme from the server value once; client toggles own it after
	setTheme(untrack(() => data.theme));
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}

<Toaster position="top-right" />
