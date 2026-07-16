<script lang="ts">
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import {
		daySummary,
		initials,
		layerName,
		personTone,
		PERSON_CLASS,
		type DaySummary,
		type Layer,
		type Schedule
	} from '$lib/oncall';

	let { layers, from }: { layers: Layer[]; from: string } = $props();

	const DAY = 86_400_000;
	const WEEKDAY = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

	const days = $derived(
		Array.from({ length: 7 }, (_, index) => new Date(Date.parse(`${from}T00:00:00Z`) + index * DAY))
	);

	const draft = $derived<Schedule>({
		id: 'draft',
		name: 'draft',
		team: '',
		layers,
		overrides: [],
		audit: [],
		feedToken: '',
		archived: false,
		paused: false
	});

	const effective = $derived(days.map((day) => daySummary(draft, day)));

	const rows = $derived(
		layers.map((layer, index) => ({
			label: `L${layers.length - index}`,
			title: layerName(layers.length, index),
			days: days.map((day) => daySummary({ ...draft, layers: [layer] }, day))
		}))
	);
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
	{#each days as day (day.toISOString())}
		<span
			class="text-subtle-foreground text-center text-[9.5px] leading-[1.5] tracking-[0.05em] uppercase"
		>
			{WEEKDAY[day.getUTCDay()]}<br />{day.getUTCDate()}
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
