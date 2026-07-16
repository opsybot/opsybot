<script lang="ts">
	import Page from '$lib/components/layout/page.svelte';
	import Builder from '$lib/components/workflows/builder.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="Workflows" subtitle="Automate the boring half of incident response">
	<div class="mt-[18px]">
		<!-- {#key} reseeds the builder so seeded state cannot outlive a navigation -->
		{#key data}
			<Builder
				initial={data.initial}
				editing={false}
				fromTemplate={data.fromTemplate}
				roleNames={data.roleNames}
			/>
		{/key}
	</div>
</Page>
