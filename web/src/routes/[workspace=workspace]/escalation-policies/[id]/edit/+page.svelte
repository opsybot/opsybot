<script lang="ts">
	import PolicyBuilder from '$lib/components/escalation/policy-builder.svelte';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<!-- {#key} reseeds the builder when the loaded policy changes -->
{#key data}
	<PolicyBuilder initial={data.tree} backHref={ws(`/escalation-policies/${data.id}`)} />
{/key}
