<script lang="ts">
	import { untrack } from 'svelte';
	import { formatAge } from '$lib/time';

	let { since, now: serverNow }: { since: string; now: number } = $props();

	let now = $state(untrack(() => serverNow));

	$effect(() => {
		const timer = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(timer);
	});
</script>

<span>{formatAge(now - Date.parse(since))}</span>
