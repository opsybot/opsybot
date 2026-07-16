<script lang="ts">
	import Page from '$lib/components/layout/page.svelte';
	import Builder from '$lib/components/workflows/builder.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="Workflows" subtitle="Automate the boring half of incident response">
	<div class="mt-[18px]">
		{#key data}
			<Builder initial={data.initial} editing={true} fromTemplate={false} roleNames={data.roleNames} />
		{/key}
	</div>
</Page>
