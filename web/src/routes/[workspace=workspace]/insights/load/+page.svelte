<script lang="ts">
	import ScaleIcon from '@lucide/svelte/icons/scale';
	import EmptyState from '$lib/components/insights/empty-state.svelte';
	import Filters from '$lib/components/insights/filters.svelte';
	import StatBar from '$lib/components/insights/stat-bar.svelte';
	import UserAvatar from '$lib/components/layout/user-avatar.svelte';
	import { COHORT_FLOOR } from '$lib/insights';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const load = $derived(data.load);
	const maxPages = $derived(load && load.rows.length ? Math.max(...load.rows.map((row) => row.pages)) : 1);
</script>

{#if !data.available || !load}
	<EmptyState />
{:else}
	<div class="flex flex-col gap-3.5">
		<Filters tab="load" />

		<div
			class="bg-brand-wash border-brand-edge text-muted-foreground flex items-center gap-2.5 rounded-md border px-3.5 py-[11px] text-[12.5px] leading-[1.5]"
		>
			<ScaleIcon class="text-brand-foreground size-3.5 shrink-0" />
			<span>{load.note}</span>
		</div>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">{load.header}</span>
			</header>

			{#if load.withheld}
				<div class="text-muted-foreground px-4 py-8 text-center text-[13px] leading-[1.55]">
					Fewer than {COHORT_FLOOR} people match this filter. On-call load isn’t shown for cohorts this
					small, so no one is singled out.
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full border-collapse text-[13px]">
						<thead>
							<tr class="text-subtle-foreground text-left text-[11px] tracking-[0.05em] uppercase">
								<th class="py-2.5 pr-3 pl-4 font-semibold">Person</th>
								<th class="px-3 py-2.5 font-semibold">On-call hours</th>
								<th class="px-3 py-2.5 font-semibold">Pages</th>
								<th class="px-3 py-2.5 font-semibold">Night pages</th>
								<th class="py-2.5 pr-4 pl-3 font-semibold">Weekend pages</th>
							</tr>
						</thead>
						<tbody>
							{#each load.rows as person (person.name)}
								<tr class="border-t" data-person={person.name}>
									<td class="py-[11px] pr-3 pl-4">
										<div class="flex items-center gap-2">
											<UserAvatar name={person.name} size="xs" />
											<span class="text-[13px]">{person.name}</span>
										</div>
									</td>
									<td class="px-3 py-[11px] font-mono text-[12.5px]">{person.hours} h</td>
									<td class="px-3 py-[11px]">
										<div class="flex items-center gap-2">
											<div class="w-[70px]"><StatBar value={person.pages} max={maxPages} /></div>
											<span class="font-mono text-[12.5px]">{person.pages}</span>
										</div>
									</td>
									<td class="px-3 py-[11px] font-mono text-[12.5px] {person.heavy ? 'text-warning-ink' : ''}">
										{person.night}
									</td>
									<td class="py-[11px] pr-4 pl-3 font-mono text-[12.5px]">{person.weekend}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				{#if load.footnote}
					<div class="text-subtle-foreground border-t px-4 py-2.5 text-[11.5px]">{load.footnote}</div>
				{/if}
			{/if}
		</section>
	</div>
{/if}
