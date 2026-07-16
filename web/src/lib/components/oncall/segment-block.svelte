<script lang="ts">
	import RepeatIcon from '@lucide/svelte/icons/repeat';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { formatSpanHours, personTone, PERSON_CLASS, type Segment, type Zone } from '$lib/oncall';

	let { segment, zone, reason }: { segment: Segment; zone: Zone; reason?: string } = $props();

	const hours = $derived(formatSpanHours(segment, zone));
	const label = $derived(
		segment.person && segment.override && reason
			? `${segment.person} — ${reason}`
			: (segment.person ?? 'No one is on call')
	);
</script>

{#if segment.person}
	<div
		title={label}
		class="flex min-w-0 items-center gap-[5px] overflow-hidden rounded-sm border px-[7px] py-1 text-[11px] {PERSON_CLASS[
			personTone(segment.person)
		]}"
	>
		<span class="truncate font-semibold">{segment.person.split(' ')[0]}</span>
		{#if segment.override}
			<RepeatIcon class="size-[10px] shrink-0" aria-label="Override" />
		{/if}
		<span class="ml-auto shrink-0 font-mono text-[9.5px] whitespace-nowrap">{hours}</span>
	</div>
{:else}
	<div
		title="No one is on call"
		class="bg-warning-wash border-warning-edge text-warning-ink flex min-w-0 items-center gap-[5px] overflow-hidden rounded-sm border px-[7px] py-1 text-[11px]"
	>
		<TriangleAlertIcon class="size-[11px] shrink-0" />
		<span class="truncate">gap · {hours}</span>
	</div>
{/if}
