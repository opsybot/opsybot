<script lang="ts">
	import EmptyState from '$lib/components/insights/empty-state.svelte';
	import Filters from '$lib/components/insights/filters.svelte';
	import StatBar from '$lib/components/insights/stat-bar.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { barTone } from '$lib/insights';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

{#if !data.available || !data.followups}
	<EmptyState />
{:else}
	<div class="flex flex-col gap-3.5">
		<Filters tab="followups" />

		<div class="grid grid-cols-1 gap-3 min-[800px]:grid-cols-3">
			{#each data.followups.stats as stat (stat.key)}
				<div class="bg-card overflow-hidden rounded-xl border p-4">
					<div
						class="text-[30px] font-light tracking-[-0.02em] {stat.tone === 'warning'
							? 'text-warning-ink'
							: ''}"
					>
						{stat.value}
					</div>
					<div class="mt-0.5 text-[12.5px] font-medium">{stat.key}</div>
				</div>
			{/each}
		</div>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Completion by team</span>
			</header>
			{#each data.followups.byTeam as team (team.team)}
				<div class="flex items-center gap-3 border-t px-4 py-[11px] first:border-t-0">
					<Tag>{team.team}</Tag>
					<div class="flex-1"><StatBar value={team.pct} tone={barTone(team.pct)} /></div>
					<span class="w-10 text-right font-mono text-[12.5px]">{team.pct}%</span>
				</div>
			{/each}
		</section>
	</div>
{/if}
