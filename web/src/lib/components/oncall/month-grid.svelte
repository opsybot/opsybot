<script lang="ts">
	import RepeatIcon from '@lucide/svelte/icons/repeat';
	import { initials, personTone, PERSON_CLASS, type DaySummary } from '$lib/oncall';

	let {
		days,
		blanks,
		today
	}: {
		days: DaySummary[];
		blanks: number;
		today: string;
	} = $props();

	const DOW = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
</script>

<div class="grid grid-cols-7 gap-1">
	{#each DOW as day (day)}
		<span class="text-subtle-foreground py-1.5 text-center text-[9.5px] tracking-[0.06em] uppercase">
			{day}
		</span>
	{/each}

	{#each { length: blanks }, index (index)}
		<span class="min-h-[56px] rounded-sm border border-transparent"></span>
	{/each}

	{#each days as day (day.date)}
		<span
			class="bg-inset flex min-h-[56px] flex-col items-center gap-[3px] rounded-sm border px-0.5 py-1.5"
			style={day.date === today ? 'box-shadow: var(--focus-ring)' : undefined}
		>
			<span class="text-subtle-foreground font-mono text-[10.5px]">
				{Number(day.date.slice(8))}
			</span>

			{#if day.person}
				<span
					title={day.person}
					class="inline-flex items-center gap-[3px] rounded-full border px-[7px] py-0.5 text-[10.5px] font-semibold {PERSON_CLASS[
						personTone(day.person)
					]}"
				>
					{initials(day.person)}
					{#if day.override}
						<RepeatIcon class="size-[9px]" aria-label="Override" />
					{/if}
				</span>
			{/if}

			{#if day.gap}
				<span class="text-warning-ink font-mono text-[9px]">gap</span>
			{/if}
		</span>
	{/each}
</div>
