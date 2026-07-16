<script lang="ts">
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import { Button } from '$lib/components/ui/button';
	import type { Shift } from '$lib/dashboard';
	import { untrack } from 'svelte';
	import { ws } from '$lib/navigation';
	import { formatUtcTime } from '$lib/time';

	let { shift, now: serverNow }: { shift: Shift; now: number } = $props();

	let now = $state(untrack(() => serverNow));

	$effect(() => {
		const timer = setInterval(() => (now = Date.now()), 30_000);
		return () => clearInterval(timer);
	});

	const start = $derived(Date.parse(shift.start));
	const end = $derived(Date.parse(shift.end));
	// Floor at 0.04 keeps a sliver of bar visible before the shift starts
	const progress = $derived(Math.min(1, Math.max(0.04, (now - start) / (end - start))));
</script>

<div class="bg-brand-wash border-brand-edge flex items-center gap-3.5 rounded-xl border px-4 py-[13px]">
	<span
		class="bg-primary shadow-glow motion-safe:animate-pulse-brand size-2.5 shrink-0 rounded-full"
		aria-hidden="true"
	></span>

	<div class="min-w-0 flex-1">
		<div class="text-brand-foreground text-[13.5px] font-semibold">You're on call</div>
		<div class="text-muted-foreground mt-px mb-2 text-[12.5px]">
			{shift.team} rotation · SEV1 and SEV2 pages come to you first.
		</div>

		<div class="bg-inset h-1 max-w-[420px] overflow-hidden rounded-full" aria-hidden="true">
			<span
				class="bg-primary block h-full rounded-full transition-[width] duration-[400ms] ease-out"
				style="width: {(progress * 100).toFixed(1)}%"
			></span>
		</div>
		<div
			class="text-subtle-foreground mt-1 flex max-w-[420px] justify-between font-mono text-[10.5px]"
		>
			<span>{formatUtcTime(shift.start).replace(' UTC', '')}</span>
			<span>{formatUtcTime(shift.end)}</span>
		</div>
	</div>

	<Button variant="secondary" size="sm" href={ws('/on-call')}>
		<CalendarClockIcon data-icon="inline-start" />
		View schedule
	</Button>
</div>
