<script lang="ts">
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import PlayIcon from '@lucide/svelte/icons/play';
	import RepeatIcon from '@lucide/svelte/icons/repeat';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { Alert, AlertContent, AlertTitle } from '$lib/components/ui/alert';
	import Page from '$lib/components/layout/page.svelte';
	import ArchiveDialog from '$lib/components/oncall/archive-dialog.svelte';
	import AuditTrail from '$lib/components/oncall/audit-trail.svelte';
	import FeedCard from '$lib/components/oncall/feed-card.svelte';
	import HandoversCard from '$lib/components/oncall/handovers-card.svelte';
	import MonthGrid from '$lib/components/oncall/month-grid.svelte';
	import OverrideDialog from '$lib/components/oncall/override-dialog.svelte';
	import ResolverCard from '$lib/components/oncall/resolver-card.svelte';
	import WeekGrid from '$lib/components/oncall/week-grid.svelte';
	import Segmented from '$lib/components/segmented.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	let overriding = $state(false);
	let archiving = $state(false);

	// Local zone is browser-only; SSR renders UTC to avoid hydration mismatch
	let mounted = $state(false);
	let localZone = $state('');

	onMount(() => {
		mounted = true;
		const parts = new Intl.DateTimeFormat(undefined, { timeZoneName: 'short' }).formatToParts(
			new Date()
		);
		localZone = parts.find((part) => part.type === 'timeZoneName')?.value ?? '';
	});

	const zone = $derived(data.zone === 'local' && mounted ? 'local' : 'utc');

	function withParams(changes: Record<string, string>): string {
		const url = new URL(page.url);
		for (const [key, value] of Object.entries(changes)) url.searchParams.set(key, value);
		return url.pathname + url.search;
	}

	const today = $derived(new Date(data.now).toISOString().slice(0, 10));
</script>

<Page title="On-call" subtitle="Schedules, overrides, and who is next">
	<div class="flex flex-col gap-3.5">
		<a
			href={ws('/on-call')}
			class="text-muted-foreground hover:text-brand-foreground inline-flex items-center gap-1.5 self-start text-[12.5px]"
		>
			<ArrowLeftIcon class="size-3.5" />
			All schedules
		</a>

		<div class="flex flex-wrap items-center gap-2.5">
			<h2 class="m-0 font-mono text-[18px] font-semibold">{data.name}</h2>
			<Tag>{data.team}</Tag>
			{#if data.gap}
				<Badge tone="warning" size="sm" dot>{data.gap}</Badge>
			{/if}
			{#if data.paused}
				<Badge tone="neutral" size="sm">paused</Badge>
			{/if}
			{#if data.archived}
				<Badge tone="neutral" size="sm">archived</Badge>
			{/if}

			<div class="flex-1"></div>

			{#if !data.archived}
				{#if data.paused}
					<form method="POST" action="?/resume" use:enhance>
						<Button type="submit" size="sm">
							<PlayIcon data-icon="inline-start" />
							Resume
						</Button>
					</form>
				{/if}

				<Button size="sm" variant="secondary" onclick={() => (overriding = true)}>
					<RepeatIcon data-icon="inline-start" />
					Add override
				</Button>
				<Button size="sm" variant="secondary" href={ws(`/on-call/${data.id}/edit`)}>
					<PencilIcon data-icon="inline-start" />
					Edit
				</Button>

				<form
					method="POST"
					action="?/duplicate"
					use:enhance={() => {
						// Capture the name before the redirect swaps data to the copy
						const from = data.name;
						return async ({ result, update }) => {
							await update();
							if (result.type === 'redirect') {
								toast.success(`Duplicated from ${from}. The copy starts paused.`);
							}
						};
					}}
				>
					<Button type="submit" size="sm" variant="ghost">
						<CopyIcon data-icon="inline-start" />
						Duplicate
					</Button>
				</form>

				<Button size="sm" variant="ghost" onclick={() => (archiving = true)}>
					<ArchiveIcon data-icon="inline-start" />
					Archive
				</Button>
			{/if}
		</div>

		{#if data.archived}
			<Alert tone="neutral">
				<ArchiveIcon />
				<AlertContent>
					<AlertTitle>This schedule is archived</AlertTitle>
					It pages no one. The calendar below is what it used to say, kept so the history still reads.
				</AlertContent>
			</Alert>
		{:else if data.paused}
			<Alert tone="warning">
				<TriangleAlertIcon />
				<AlertContent>
					<AlertTitle>This schedule is paused</AlertTitle>
					It pages no one, so it carries no shifts and raises no gaps. The calendar below is what it
					would say once it is resumed.
				</AlertContent>
			</Alert>
		{/if}

		<div class="grid items-start gap-3.5 min-[1100px]:grid-cols-[minmax(0,1fr)_300px]">
			<section class="bg-card min-w-0 overflow-hidden rounded-xl border">
				<header class="flex flex-wrap items-center gap-2.5 border-b px-4 py-3">
					<Segmented
						label="Calendar range"
						current={data.view}
						options={[
							{ value: 'week', label: 'Week', href: withParams({ view: 'week' }) },
							{ value: 'month', label: 'Month', href: withParams({ view: 'month' }) }
						]}
					/>
					<span class="text-subtle-foreground text-[12.5px]">
						{data.view === 'week' ? data.weekLabel : data.month.label}
					</span>

					<div class="flex-1"></div>

					<Segmented
						label="Timezone"
						current={data.zone}
						options={[
							{ value: 'utc', label: 'UTC', href: withParams({ tz: 'utc' }) },
							{
								value: 'local',
								label: localZone ? `Local (${localZone})` : 'Local',
								href: withParams({ tz: 'local' })
							}
						]}
					/>
				</header>

				<div class="overflow-x-auto p-3.5">
					{#if data.view === 'week'}
						<WeekGrid
							days={data.days}
							effective={data.effective}
							layers={data.layers}
							reasons={data.reasons}
							{zone}
						/>
					{:else}
						<MonthGrid days={data.month.days} blanks={data.month.blanks} {today} />
					{/if}
				</div>

				<div class="text-subtle-foreground flex gap-4 border-t px-4 py-2.5 text-[11px]">
					<span class="inline-flex items-center gap-[5px]">
						<RepeatIcon class="size-[11px]" />
						override
					</span>
					<span class="text-warning-ink inline-flex items-center gap-[5px]">
						<TriangleAlertIcon class="size-[11px]" />
						gap — no coverage
					</span>
					<span class="ml-auto">higher layers take precedence</span>
				</div>
			</section>

			<div class="flex flex-col gap-3.5">
				<HandoversCard handovers={data.handovers} now={data.now} />
				<ResolverCard
					date={data.resolver.date}
					time={data.resolver.time}
					coverage={data.resolver.coverage}
					keep={{ view: data.view, tz: data.zone }}
				/>
				<FeedCard url={data.feedUrl} />
			</div>
		</div>

		<AuditTrail entries={data.audit} />
	</div>
</Page>

<OverrideDialog
	bind:open={overriding}
	schedule={data.name}
	target={data.target}
	people={data.people}
	now={data.now}
	error={form?.error}
/>

<ArchiveDialog bind:open={archiving} name={data.name} />
