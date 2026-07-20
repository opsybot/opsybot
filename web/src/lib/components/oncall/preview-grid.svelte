<script lang="ts">
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { initials, personTone, PERSON_CLASS, type DaySummary } from '$lib/oncall';

	let {
		days,
		effective,
		rows
	}: {
		days: { label: string; num: number; iso: string }[];
		effective: DaySummary[];
		rows: { label: string; title: string; days: DaySummary[] }[];
	} = $props();
</script>

{#snippet chip(day: DaySummary, holes = false)}
	{@const uncovered = holes && day.gap}
	{#if day.person}
		<span
			title={uncovered ? `${day.person} — part of this day is uncovered` : day.person}
			class="flex h-6 items-center justify-center gap-0.5 rounded-sm border text-[10.5px] font-semibold {uncovered
				? 'bg-warning-wash border-warning-edge text-warning-ink'
				: PERSON_CLASS[personTone(day.person)]}"
		>
			{initials(day.person)}
			{#if uncovered}
				<TriangleAlertIcon class="size-[9px]" />
			{/if}
		</span>
	{:else}
		<span
			title="No one is on call"
			class="h-[22px] rounded-sm opacity-50"
			style="background: repeating-linear-gradient(-45deg, transparent 0 5px, var(--border) 5px 6px)"
		></span>
	{/if}
{/snippet}

<div class="grid grid-cols-[52px_repeat(7,1fr)] items-center gap-1">
	<span></span>
	{#each days as day (day.iso)}
		<span
			class="text-subtle-foreground text-center text-[9.5px] leading-[1.5] tracking-[0.05em] uppercase"
		>
			{day.label}<br />{day.num}
		</span>
	{/each}

	<span
		class="text-muted-foreground pr-1.5 text-right text-[10px] tracking-[0.07em] uppercase"
		title="Who actually gets paged"
	>
		Effective
	</span>
	{#each effective as day (day.date)}
		{@render chip(day, true)}
	{/each}

	{#each rows as row (row.label)}
		<span
			class="text-muted-foreground pr-1.5 text-right text-[10px] tracking-[0.07em] uppercase"
			title={row.title}
		>
			{row.label}
		</span>
		{#each row.days as day (day.date)}
			{@render chip(day)}
		{/each}
	{/each}
</div>
