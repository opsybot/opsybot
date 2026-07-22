<script lang="ts">
	import BellOffIcon from '@lucide/svelte/icons/bell-off';
	import CheckIcon from '@lucide/svelte/icons/check';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import { SvelteSet } from 'svelte/reactivity';
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import AlertsTable from '$lib/components/alerts/alerts-table.svelte';
	import AlertsTabs from '$lib/components/alerts/alerts-tabs.svelte';
	import FilterBar from '$lib/components/alerts/filter-bar.svelte';
	import Page from '$lib/components/layout/page.svelte';
	import { Button } from '$lib/components/ui/button';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const selected = new SvelteSet<string>();

	const olderHref = $derived.by(() => {
		if (!data.nextCursor) return null;
		const next = new URLSearchParams(page.url.searchParams);
		next.set('cursor', data.nextCursor);
		return `?${next}`;
	});

	const newestHref = $derived.by(() => {
		const next = new URLSearchParams(page.url.searchParams);
		next.delete('cursor');
		return `?${next}`;
	});
</script>

<Page title="Alerts" subtitle="Deduplicated signals from every connected source">
	<AlertsTabs current="list" />

	{#if data.alerts.length === 0 && (data.filtered || data.paged)}
		<FilterBar sources={data.sources} services={data.services} labels={data.labels} />
		<div class="bg-card flex flex-col items-center gap-2.5 rounded-xl border px-5 py-13">
			<div class="text-[15px] font-medium">Nothing matches these filters</div>
			<Button variant="secondary" size="sm" href={ws('/alerts')}>Reset all filters</Button>
		</div>
	{:else if data.alerts.length === 0}
		<div class="bg-card flex flex-col items-center gap-2.5 rounded-xl border px-5 py-13">
			<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
				<BellOffIcon class="text-subtle-foreground size-5" />
			</span>
			<div class="text-[15px] font-medium">No alerts</div>
			<p class="text-subtle-foreground m-0 max-w-[400px] text-center text-[13px] leading-[1.55]">
				Alerts from connected sources land here, grouped and deduplicated. Connect a source to see
				them.
			</p>
			<Button variant="secondary" size="sm" href={ws('/alert-sources')}>Connect a source</Button>
		</div>
	{:else}
		<FilterBar sources={data.sources} services={data.services} labels={data.labels} />

		{#if selected.size > 0}
			<form
				method="POST"
				action="?/bulk"
				use:enhance={() => async ({ update }) => {
					await update();
					selected.clear();
				}}
				class="bg-brand-wash border-brand-edge flex items-center gap-2.5 rounded-md border px-3.5 py-2"
			>
				{#each selected as id (id)}
					<input type="hidden" name="id" value={id} />
				{/each}

				<span class="text-[13px] font-medium">{selected.size} selected</span>

				<Button type="submit" name="status" value="acked" variant="secondary" size="sm">
					<CheckIcon data-icon="inline-start" />
					Ack
				</Button>
				<Button type="submit" name="status" value="resolved" variant="secondary" size="sm">
					<CircleCheckIcon data-icon="inline-start" />
					Resolve
				</Button>

				<button
					type="button"
					onclick={() => selected.clear()}
					class="text-muted-foreground hover:text-brand-foreground text-[12.5px]"
				>
					Clear selection
				</button>
			</form>
		{/if}

		<div class="bg-card overflow-hidden rounded-xl border">
			<AlertsTable alerts={data.alerts} now={data.now} {selected} />
		</div>

		{#if olderHref || data.paged}
			<nav class="flex items-center gap-2.5" aria-label="Alert pages">
				{#if data.paged}
					<Button variant="secondary" size="sm" href={newestHref}>Back to newest</Button>
				{/if}
				<div class="flex-1"></div>
				{#if olderHref}
					<Button variant="secondary" size="sm" href={olderHref}>Older alerts</Button>
				{/if}
			</nav>
		{/if}
	{/if}
</Page>
