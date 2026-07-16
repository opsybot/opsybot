<script lang="ts">
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import RepeatIcon from '@lucide/svelte/icons/repeat';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import Page from '$lib/components/layout/page.svelte';
	import FilterBar from '$lib/components/postmortems/filter-bar.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { ws } from '$lib/navigation';
	import { formatUtcDate } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<Page title="Postmortems" subtitle="Blameless, drafted from the timeline">
	<div class="grid items-start gap-3.5 min-[1100px]:grid-cols-[minmax(0,1fr)_300px]">
		<div class="flex min-w-0 flex-col gap-3.5">
			{#if data.anyPublished}
				<FilterBar
					query={data.filters.query}
					service={data.filters.service}
					severity={data.filters.severity}
					range={data.filters.range}
				/>
			{/if}

			{#if !data.anyPublished}
				<div
					class="text-muted-foreground flex flex-col items-center gap-2.5 rounded-xl border border-dashed px-5 py-14"
				>
					<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
						<FileTextIcon class="text-subtle-foreground size-5" />
					</span>
					<div class="text-[15px] font-medium">No postmortems yet</div>
					<p class="text-subtle-foreground m-0 max-w-[420px] text-center text-[13px] leading-[1.55]">
						After an incident resolves, Opsybot drafts the postmortem from its timeline. The library
						builds itself as you close incidents.
					</p>
				</div>
			{:else}
				<div class="bg-card overflow-hidden rounded-xl border">
					{#each data.library as entry (entry.id)}
						<a
							href={ws(`/postmortems/${entry.id}`)}
							data-postmortem={entry.id}
							class="hover:bg-accent flex items-center gap-3 border-t px-4 py-3 first:border-t-0"
						>
							<Badge tone={entry.tone} size="sm">{entry.severity}</Badge>

							<div class="min-w-0 flex-1">
								<div class="text-foreground text-[13.5px] font-medium">{entry.title}</div>
								<div class="text-subtle-foreground mt-0.5 font-mono text-[11px]">
									{entry.id} · {entry.incidentId} · published {formatUtcDate(entry.publishedAt)} · {entry.author}
								</div>
							</div>

							<div class="flex max-w-[260px] flex-wrap justify-end gap-[5px]">
								{#each entry.services as service (service)}
									<Tag>{service}</Tag>
								{/each}
							</div>

							<ChevronRightIcon class="text-subtle-foreground size-[15px] shrink-0" />
						</a>
					{:else}
						<div class="text-subtle-foreground px-4 py-[30px] text-center text-[13px]">
							Nothing matches these filters.
						</div>
					{/each}
				</div>

				{#if data.waiting.length}
					<section class="bg-card overflow-hidden rounded-xl border">
						<header class="flex flex-wrap items-center gap-2 border-b px-4 py-3">
							<ClockIcon class="text-warning-ink size-3.5" />
							<span class="text-[13.5px] font-semibold">Waiting for a postmortem</span>
							<span class="text-subtle-foreground ml-1 text-xs">
								SEV1–SEV2 incidents require one within 3 working days
							</span>
						</header>

						{#each data.waiting as entry (entry.incidentId)}
							<div
								data-waiting={entry.incidentId}
								class="flex items-center gap-3 border-t px-4 py-3 first:border-t-0"
							>
								<Badge tone={entry.tone} size="sm">{entry.severity}</Badge>

								<div class="min-w-0 flex-1">
									<div class="text-[13.5px] font-medium">{entry.title}</div>
									<div class="text-subtle-foreground mt-0.5 font-mono text-[11px]">
										{entry.incidentId} · {entry.resolved}
									</div>
								</div>

								{#if entry.overdue}
									<Badge tone="critical" size="sm" dot>{entry.state}</Badge>
									<Button size="sm" variant="secondary" href={ws(`/postmortems/${entry.postmortemId}`)}>
										<SparklesIcon data-icon="inline-start" />
										Start draft
									</Button>
								{:else}
									<Badge tone="neutral" size="sm">{entry.state}</Badge>
								{/if}
							</div>
						{/each}
					</section>
				{/if}
			{/if}
		</div>

		{#if data.anyPublished}
			<section class="bg-card self-start overflow-hidden rounded-xl border">
				<header class="flex items-center gap-2 border-b px-4 py-3">
					<RepeatIcon class="text-subtle-foreground size-3.5" />
					<span class="text-[13.5px] font-semibold">Recurring contributing factors</span>
				</header>

				<div class="py-1">
					{#each data.patterns as pattern (pattern.label)}
						<div class="border-t px-3.5 py-[11px] first:border-t-0">
							<div class="flex items-center gap-2">
								<span class="flex-1 text-[13px] font-medium">{pattern.label}</span>
								<Badge tone="warning" size="sm">×{pattern.count}</Badge>
							</div>
							<div class="text-subtle-foreground mt-[3px] font-mono text-[10.5px]">
								{pattern.postmortems.join(' · ')}
							</div>
						</div>
					{:else}
						<p class="text-subtle-foreground m-0 px-3.5 py-[11px] text-[12.5px]">
							No condition has caused two incidents yet.
						</p>
					{/each}
				</div>

				<p
					class="text-subtle-foreground m-0 border-t px-3.5 py-2.5 text-[11.5px] leading-[1.5]"
				>
					Patterns across published postmortems. Two incidents from the same missing canary stage is
					a signal, not a coincidence.
				</p>
			</section>
		{/if}
	</div>
</Page>
