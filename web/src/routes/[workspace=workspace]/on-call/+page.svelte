<script lang="ts">
	import CalendarClockIcon from '@lucide/svelte/icons/calendar-clock';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import Page from '$lib/components/layout/page.svelte';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import OncallTabs from '$lib/components/oncall/oncall-tabs.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { ws } from '$lib/navigation';
	import { formatWhen } from '$lib/oncall';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="On-call" subtitle="Schedules, overrides, and who is next">
	<OncallTabs current="schedules" />

	<div class="flex flex-col gap-3.5">
		<div class="flex items-center">
			<span class="text-subtle-foreground text-[13px]">
				{data.schedules.length}
				{data.schedules.length === 1 ? 'schedule' : 'schedules'}
			</span>
			<div class="flex-1"></div>
			<Button size="sm" href={ws('/on-call/new')}>
				<PlusIcon data-icon="inline-start" />
				New schedule
			</Button>
		</div>

		{#if data.schedules.length === 0}
			<div
				class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
			>
				<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
					<CalendarClockIcon class="text-subtle-foreground size-5" />
				</span>
				<div class="text-[15px] font-medium">No schedules</div>
				<p class="text-subtle-foreground m-0 max-w-[400px] text-center text-[13px] leading-[1.55]">
					A schedule decides who gets paged, and when. Most teams start with one weekly rotation.
				</p>
				<Button variant="secondary" size="sm" href={ws('/on-call/new')}>
					<PlusIcon data-icon="inline-start" />
					Create your first schedule
				</Button>
			</div>
		{:else}
			<div class="bg-card overflow-hidden rounded-xl border">
				{#each data.schedules as schedule (schedule.id)}
					<a
						href={ws(`/on-call/${schedule.id}`)}
						data-schedule={schedule.id}
						class="hover:bg-accent flex items-center gap-[18px] border-t px-4 py-3.5 first:border-t-0"
					>
						<div class="min-w-0 flex-1">
							<div class="flex flex-wrap items-center gap-2">
								<span class="text-foreground font-mono text-[13.5px] font-medium">
									{schedule.name}
								</span>
								<Tag>{schedule.team}</Tag>
								{#if schedule.gap}
									<Badge tone="warning" size="sm" dot>{schedule.gap}</Badge>
								{/if}
								{#if schedule.paused}
									<Badge tone="neutral" size="sm">paused</Badge>
								{/if}
							</div>
							<div class="text-subtle-foreground mt-1 font-mono text-[11.5px]">
								{#if schedule.paused}
									paused — it pages no one
								{:else if schedule.handover}
									next handover {formatWhen(schedule.handover.at, data.now)} → {schedule.handover.to}
								{:else}
									no handover in the next two weeks
								{/if}
							</div>
						</div>

						<div class="flex items-center gap-2">
							{#if schedule.person && schedule.until}
								<UserAvatar name={schedule.person} size="xs" onCall />
								<div>
									<div class="text-[12.5px] font-medium">{schedule.person}</div>
									<div class="text-subtle-foreground font-mono text-[10.5px]">
										on call until {formatWhen(schedule.until, data.now)}
									</div>
								</div>
							{:else if !schedule.paused}
								<Badge tone="warning" size="sm" dot>nobody on call</Badge>
							{/if}
						</div>

						<ChevronRightIcon class="text-subtle-foreground size-4 shrink-0" />
					</a>
				{/each}
			</div>
		{/if}
	</div>
</Page>
