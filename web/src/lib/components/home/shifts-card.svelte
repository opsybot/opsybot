<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import type { Shift } from '$lib/dashboard';
	import { ws } from '$lib/navigation';
	import { formatUtcDate, formatUtcTime } from '$lib/time';
	import RailCard from './rail-card.svelte';

	let { shifts, now }: { shifts: Shift[]; now: number } = $props();

	const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

	const today = $derived(formatUtcDate(new Date(now).toISOString()));

	const week = $derived.by(() =>
		Array.from({ length: 7 }, (_, offset) => {
			const day = new Date(now + offset * 86_400_000);
			const date = formatUtcDate(day.toISOString());
			return {
				date,
				letter: WEEKDAYS[day.getUTCDay()][0],
				number: day.getUTCDate(),
				today: offset === 0,
				covered: shifts.some((shift) => formatUtcDate(shift.start) === date)
			};
		})
	);

	function label(shift: Shift): string {
		const date = formatUtcDate(shift.start);
		if (date === today) return 'Today';
		return `${WEEKDAYS[new Date(shift.start).getUTCDay()]} ${date}`;
	}

	function isNow(shift: Shift): boolean {
		return now >= Date.parse(shift.start) && now < Date.parse(shift.end);
	}
</script>

<RailCard title="My next 7 days">
	{#snippet footer()}
		<a href={ws('/on-call')} class="text-brand-foreground hover:underline">Full schedule</a>
	{/snippet}

	<div class="grid grid-cols-7 gap-[5px] px-4 pt-3 pb-1" aria-hidden="true">
		{#each week as day (day.date)}
			<span
				class="bg-inset flex flex-col items-center gap-px rounded-sm border pt-[5px] pb-1 {day.covered
					? 'bg-brand-wash border-brand-edge'
					: ''}"
				style={day.today
					? 'box-shadow: 0 0 0 3px color-mix(in srgb, var(--primary) 35%, transparent)'
					: undefined}
			>
				<span
					class="text-[9.5px] tracking-[0.06em] uppercase {day.covered
						? 'text-brand-foreground'
						: 'text-subtle-foreground'}"
				>
					{day.letter}
				</span>
				<span
					class="font-mono text-[11.5px] {day.covered ? 'text-foreground' : 'text-muted-foreground'}"
				>
					{day.number}
				</span>
			</span>
		{/each}
	</div>

	{#each shifts as shift (shift.start)}
		<div class="flex items-center gap-2.5 border-t px-4 py-2.5">
			<span
				class="size-2 shrink-0 rounded-full {isNow(shift)
					? 'bg-primary shadow-glow'
					: 'bg-border-strong'}"
				aria-hidden="true"
			></span>
			<div class="min-w-0 flex-1">
				<div class="flex items-center gap-1.5 text-[13px] font-medium">
					{label(shift)}
					{#if isNow(shift)}
						<Badge tone="brand" size="sm">now</Badge>
					{/if}
				</div>
				<div class="text-subtle-foreground mt-px font-mono text-[11px]">
					{formatUtcTime(shift.start).replace(' UTC', '')}–{formatUtcTime(shift.end)} · {shift.team}
				</div>
			</div>
		</div>
	{/each}
</RailCard>
