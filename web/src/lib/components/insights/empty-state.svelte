<script lang="ts">
	import ChartLineIcon from '@lucide/svelte/icons/chart-line';
	import * as Empty from '$lib/components/ui/empty';
</script>

<Empty.Root class="border">
	<Empty.Header>
		<Empty.Media variant="icon">
			<ChartLineIcon />
		</Empty.Media>
		<Empty.Title>Nothing to measure yet</Empty.Title>
		<Empty.Description>
			Every metric computes from incident timelines, so the numbers fill in as incidents are resolved.
		</Empty.Description>
	</Empty.Header>
</Empty.Root>
