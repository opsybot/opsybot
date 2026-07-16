<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div class="flex max-w-[760px] flex-col gap-3.5">
	<p class="text-muted-foreground m-0 text-[13px] leading-[1.6]">
		Every metric, defined precisely. The same definitions power the inline (i) tooltips across
		Insights.
	</p>

	<div class="bg-card overflow-hidden rounded-xl border">
		{#each data.definitions as entry (entry.term)}
			<div class="border-t px-4 py-[13px] first:border-t-0">
				<div class="mb-[3px] text-[13px] font-semibold">{entry.term}</div>
				<div class="text-muted-foreground text-[12.5px] leading-[1.55]">{entry.definition}</div>
			</div>
		{/each}
	</div>
</div>
