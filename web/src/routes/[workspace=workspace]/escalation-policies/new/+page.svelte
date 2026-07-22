<script lang="ts">
	import PolicyBuilder from '$lib/components/escalation/policy-builder.svelte';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#key data}
	<PolicyBuilder initial={data.starter} directory={data.directory} backHref={ws('/escalation-policies')} />
{/key}
