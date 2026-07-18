<script lang="ts">
	import '../app.css';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import favicon from '$lib/assets/favicon.svg';
	import { Toaster } from '$lib/components/ui/sonner';
	import { takeFlashCookie } from '$lib/flash';
	import { setTheme } from '$lib/theme.svelte';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	// Seed theme from the server value once; client toggles own it after
	setTheme(untrack(() => data.theme));

	$effect(() => {
		const flash = takeFlashCookie();
		if (!flash) return;
		const notify = flash.tone === 'error' ? toast.error : flash.tone === 'info' ? toast.info : toast.success;
		notify(flash.title, { description: flash.message });
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}

<Toaster position="top-right" />
