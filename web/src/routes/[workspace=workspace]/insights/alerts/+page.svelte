<script lang="ts">
	import Chart from '$lib/components/insights/chart.svelte';
	import EmptyState from '$lib/components/insights/empty-state.svelte';
	import Filters from '$lib/components/insights/filters.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#if !data.available || !data.alerts}
	<EmptyState />
{:else}
	<div class="flex flex-col gap-3.5">
		<Filters tab="alerts" />

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Alert volume</span>
				<span class="text-subtle-foreground ml-auto font-mono text-[11px]">per day · last 14 days</span>
			</header>
			<div class="p-4">
				<Chart type="bar" tone="info" height={130} data={data.alerts.volume} labels={['14d ago', 'today']} />
			</div>
		</section>

		<div class="grid grid-cols-1 gap-3 min-[800px]:grid-cols-3">
			{#each data.alerts.stats as stat (stat.key)}
				<div class="bg-card overflow-hidden rounded-xl border p-4">
					<div class="text-[30px] font-light tracking-[-0.02em]">{stat.value}</div>
					<div class="mt-0.5 text-[12.5px] font-medium">{stat.key}</div>
					<div class="text-subtle-foreground mt-[3px] text-[11.5px] leading-[1.4]">{stat.note}</div>
				</div>
			{/each}
		</div>
	</div>
{/if}
