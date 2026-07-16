<script lang="ts">
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import RepeatIcon from '@lucide/svelte/icons/repeat';
	import { toast } from 'svelte-sonner';
	import Page from '$lib/components/layout/page.svelte';
	import OncallTabs from '$lib/components/oncall/oncall-tabs.svelte';
	import SwapDialog from '$lib/components/oncall/swap-dialog.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { formatRemaining, formatUtcTime } from '$lib/time';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	let swapping = $state<{ when: string; schedule: string } | null>(null);
	let open = $state(false);

	function requestSwap(shift: { when: string; schedule: string }) {
		swapping = shift;
		open = true;
	}

	$effect(() => {
		if (form?.person) {
			toast.success(`Swap requested. ${form.person} gets a notification to accept.`);
		}
	});

	const days = $derived(new Set(data.month.days));
</script>

<Page title="On-call" subtitle="Schedules, overrides, and who is next">
	<OncallTabs current="mine" />

	<div class="grid items-start gap-3.5 min-[1100px]:grid-cols-[minmax(0,1fr)_300px]">
		<div class="flex min-w-0 flex-col gap-3.5">
			<section class="bg-card overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2.5 border-b px-4 py-3">
					<span class="text-sm font-semibold">My shifts</span>
					<span class="text-subtle-foreground ml-1 text-[12.5px]">all schedules · next 7 days</span>
				</header>

				{#each data.shifts as shift (shift.id)}
					<div class="flex items-center gap-2.5 border-t px-3.5 py-2.5 first:border-t-0">
						<span
							class="bg-primary size-2 shrink-0 rounded-full"
							style="box-shadow: var(--glow-brand)"
							aria-hidden="true"
						></span>
						<div class="min-w-0 flex-1">
							<div class="text-[13px] font-medium">{shift.when}</div>
							<div class="text-subtle-foreground mt-px font-mono text-[11px]">
								{shift.schedule} ·
								{#if Date.parse(shift.startsAt) <= data.now}
									on call now, until {formatUtcTime(shift.endsAt)}
								{:else}
									starts in {formatRemaining(Date.parse(shift.startsAt) - data.now)}
								{/if}
							</div>
						</div>
						<Button size="sm" variant="ghost" onclick={() => requestSwap(shift)}>
							<RepeatIcon data-icon="inline-start" />
							Request swap
						</Button>
					</div>
				{:else}
					<div class="flex flex-col items-center gap-2.5 px-5 py-8">
						<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
							<CalendarClockIcon class="text-subtle-foreground size-5" />
						</span>
						<div class="text-sm font-medium">No shifts in the next 7 days</div>
						<p class="text-subtle-foreground m-0 text-[12.5px]">
							You are not on call this week. Someone else is carrying it.
						</p>
					</div>
				{/each}
			</section>

			<section class="bg-card overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2.5 border-b px-4 py-3">
					<span class="text-sm font-semibold">My override requests</span>
				</header>

				{#each data.requests as request (request.id)}
					<div class="flex items-center gap-2.5 border-t px-3.5 py-2.5 first:border-t-0">
						{#if request.status === 'pending'}
							<ClockIcon class="text-warning-ink size-3.5 shrink-0" />
						{:else}
							<CheckIcon class="text-success-ink size-3.5 shrink-0" />
						{/if}
						<div class="min-w-0 flex-1">
							<div class="text-[13px]">{request.text}</div>
							{#if request.message}
								<div class="text-subtle-foreground mt-px text-[11.5px]">“{request.message}”</div>
							{/if}
						</div>
						<Badge tone={request.status === 'pending' ? 'warning' : 'success'} size="sm">
							{request.status}
						</Badge>
					</div>
				{:else}
					<p class="text-subtle-foreground m-0 px-3.5 py-3.5 text-[12.5px]">
						Nothing pending. Ask for a swap from a shift above.
					</p>
				{/each}
			</section>
		</div>

		<section class="bg-card self-start overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2.5 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">{data.month.label}</span>
				<span class="text-subtle-foreground ml-auto font-mono text-[10.5px]">UTC</span>
			</header>

			<div class="grid grid-cols-7 gap-[3px] px-3.5 py-3">
				{#each ['M', 'T', 'W', 'T', 'F', 'S', 'S'] as day, index (index)}
					<span
						class="text-subtle-foreground text-center text-[9.5px] tracking-[0.06em] uppercase"
					>
						{day}
					</span>
				{/each}

				{#each { length: data.month.blanks }, index (index)}
					<span></span>
				{/each}

				{#each { length: data.month.length }, index (index)}
					{@const date = `${data.month.prefix}${String(index + 1).padStart(2, '0')}`}
					<span
						class="flex h-7 items-center justify-center rounded-sm border border-transparent font-mono text-[11px]
						{days.has(date)
							? 'bg-brand-wash border-brand-edge text-brand-foreground font-semibold'
							: 'text-muted-foreground'}"
						style={date === data.month.today ? 'box-shadow: var(--focus-ring)' : undefined}
					>
						{index + 1}
					</span>
				{/each}
			</div>

			<div class="text-subtle-foreground flex gap-3 px-3.5 pt-0.5 pb-3 text-[11px]">
				<span class="inline-flex items-center gap-[5px]">
					<span class="bg-brand-wash border-brand-edge inline-block size-3 rounded-sm border"></span>
					my shift
				</span>
				<span class="inline-flex items-center gap-[5px]">
					<span
						class="inline-block size-3 rounded-sm border border-transparent"
						style="box-shadow: var(--focus-ring)"
					></span>
					today
				</span>
			</div>
		</section>
	</div>
</Page>

<SwapDialog bind:open shift={swapping} me={data.me} />
