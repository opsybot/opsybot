<script lang="ts">
	import StatusBadge from '$lib/components/status-badge.svelte';
	import type { Alert } from '$lib/alerts';

	let { alert, size = 'sm' }: { alert: Alert; size?: 'sm' | 'md' } = $props();
</script>

{#if alert.status === 'open'}
	<StatusBadge status="firing" label="open" {size} />
{:else if alert.status === 'acked'}
	<StatusBadge status="acknowledged" label="acked · {alert.ackedBy ?? 'you'}" {size} />
{:else}
	<StatusBadge status="resolved" {size} />
{/if}
