<script lang="ts">
	import Page from '$lib/components/layout/page.svelte';

	let { children } = $props();
</script>

<Page title="Escalation policies" subtitle="Who gets paged, in what order">
	{@render children()}
</Page>
