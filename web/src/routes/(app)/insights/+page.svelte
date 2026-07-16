<script lang="ts">
	import Chart from '$lib/components/insights/chart.svelte';
	import EmptyState from '$lib/components/insights/empty-state.svelte';
	import Filters from '$lib/components/insights/filters.svelte';
	import MetricTip from '$lib/components/insights/metric-tip.svelte';
	import StatBar from '$lib/components/insights/stat-bar.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { definitionBlurb } from '$lib/insights';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#if !data.available || !data.overview}
	<EmptyState />
{:else}
	<div class="flex flex-col gap-3.5">
		<Filters tab="overview" />

		<div class="grid grid-cols-1 gap-3 min-[800px]:grid-cols-3">
			{#each data.overview.metrics as metric (metric.key)}
				<div class="bg-card overflow-hidden rounded-xl border p-4">
					<div class="flex items-center gap-1.5">
						<span class="text-subtle-foreground font-mono text-[11px] tracking-[0.05em]">
							{metric.key}
						</span>
						<MetricTip def={definitionBlurb(metric.key)} />
						<div class="flex-1"></div>
						<Badge tone={metric.good ? 'success' : 'warning'} size="sm">{metric.delta}</Badge>
					</div>
					<div class="mt-2 mb-0.5 text-[30px] font-light tracking-[-0.02em]">{metric.value}</div>
					<div class="text-subtle-foreground text-[12px]">{metric.label}</div>
					<div class="text-subtle-foreground mt-1 text-[10.5px]">{data.overview.comparison}</div>
				</div>
			{/each}
		</div>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">MTTR trend</span>
				<span class="text-subtle-foreground ml-auto font-mono text-[11px]">
					minutes · last 12 weeks
				</span>
			</header>
			<div class="p-4">
				<Chart type="line" tone="brand" height={140} data={data.overview.mttrTrend} labels={['12w ago', 'now']} />
			</div>
		</section>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Time between stages</span>
				<span class="text-subtle-foreground ml-1 text-[11.5px]">median across resolved incidents</span>
			</header>
			{#each data.overview.stages as stage (stage.label)}
				<div class="flex items-center gap-3 border-t px-4 py-[11px] first:border-t-0">
					<span class="w-[220px] text-[13px]">{stage.label}</span>
					<div class="flex-1"><StatBar value={stage.pct} /></div>
					<span class="w-[60px] text-right font-mono text-[12.5px]">{stage.value}</span>
				</div>
			{/each}
		</section>
	</div>
{/if}
