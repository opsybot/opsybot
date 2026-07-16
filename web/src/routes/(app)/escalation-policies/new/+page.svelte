<script lang="ts">
	import PolicyBuilder from '$lib/components/escalation/policy-builder.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#key data}
	<PolicyBuilder initial={data.starter} backHref="/escalation-policies" />
{/key}
