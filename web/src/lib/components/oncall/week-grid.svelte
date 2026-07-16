<script lang="ts">
	import { formatLayerLine, type Layer, type Segment, type Zone } from '$lib/oncall';
	import SegmentBlock from './segment-block.svelte';

	let {
		days,
		effective,
		layers,
		reasons,
		zone
	}: {
		days: { label: string; num: number; today: boolean }[];
		effective: Segment[][];
		layers: { layer: Layer; name: string; duty: string; days: Segment[][] }[];
		reasons: Record<string, string>;
		zone: Zone;
	} = $props();

	const COLUMNS = 'grid grid-cols-[108px_repeat(7,minmax(96px,1fr))] gap-1';
</script>

<div>
	<div class="{COLUMNS} mb-1.5">
		<span></span>
		{#each days as day (day.num)}
			<span
				class="tracking-[0.05em] rounded-sm py-1 text-center text-[11px] uppercase
				{day.today ? 'text-brand-foreground bg-brand-wash' : 'text-subtle-foreground'}"
			>
				{day.label}
				<span class="font-mono text-[10.5px]">{day.num}</span>
			</span>
		{/each}
	</div>

	<div class="{COLUMNS} mb-2 border-b border-dashed pb-2">
		<span
			class="text-muted-foreground tracking-[0.07em] self-center pr-2 text-right text-[10.5px] uppercase"
		>
			Effective
		</span>
		{#each effective as day, index (index)}
			<div class="flex min-h-[30px] flex-col gap-[3px]">
				{#each day as segment (segment.startsAt)}
					<SegmentBlock {segment} {zone} reason={reasons[segment.startsAt]} />
				{/each}
			</div>
		{/each}
	</div>

	{#each layers as row (row.layer.id)}
		<div class="{COLUMNS} mb-1">
			<span
				class="text-muted-foreground tracking-[0.07em] self-center pr-2 text-right text-[10.5px] uppercase"
				title={row.duty}
			>
				{row.name}
				<span class="text-subtle-foreground block text-[9.5px] tracking-normal normal-case">
					{formatLayerLine(row.layer)}
				</span>
			</span>

			{#each row.days as day, index (index)}
				<div class="flex min-h-[30px] flex-col gap-[3px]">
					{#each day.filter((segment) => segment.person) as segment (segment.startsAt)}
						<SegmentBlock {segment} {zone} />
					{:else}
						<span
							class="min-h-[26px] flex-1 rounded-sm opacity-50"
							style="background: repeating-linear-gradient(-45deg, transparent 0 5px, var(--border) 5px 6px)"
						></span>
					{/each}
				</div>
			{/each}
		</div>
	{/each}
</div>
