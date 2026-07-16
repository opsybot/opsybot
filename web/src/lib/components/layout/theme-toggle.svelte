<script lang="ts">
	import MoonIcon from '@lucide/svelte/icons/moon';
	import SunIcon from '@lucide/svelte/icons/sun';
	import { Button } from '$lib/components/ui/button';
	import { useTheme } from '$lib/theme.svelte';

	const theme = useTheme();
</script>

<Button
	variant="ghost"
	size="icon-sm"
	onclick={() => theme.toggle()}
	aria-label={theme.current === 'dark' ? 'Switch to the light theme' : 'Switch to the dark theme'}
>
	{#if theme.current === 'dark'}
		<SunIcon />
	{:else}
		<MoonIcon />
	{/if}
</Button>
