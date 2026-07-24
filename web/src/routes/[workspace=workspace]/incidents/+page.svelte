<script lang="ts">
	import SirenIcon from '@lucide/svelte/icons/siren';
	import DeclareDialog from '$lib/components/incidents/declare-dialog.svelte';
	import FilterBar from '$lib/components/incidents/filter-bar.svelte';
	import IncidentsTable from '$lib/components/incidents/incidents-table.svelte';
	import IncidentsTabs from '$lib/components/incidents/incidents-tabs.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Page from '$lib/components/layout/page.svelte';
	import { Button } from '$lib/components/ui/button';
	import { ws } from '$lib/navigation';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let declaring = $state(false);

	$effect(() => {
		if (!page.url.searchParams.has('declare')) return;
		declaring = true;
		goto(ws('/incidents'), { replaceState: true, noScroll: true, keepFocus: true });
	});
</script>

<Page title="Incidents" subtitle="From alert to postmortem">
	<IncidentsTabs current="list" />

	{#if data.incidents.length === 0}
		<div class="bg-card flex flex-col items-center gap-2.5 rounded-xl border px-5 py-13">
			<span class="bg-inset flex size-[42px] items-center justify-center rounded-full border">
				<SirenIcon class="text-subtle-foreground size-5" />
			</span>
			<div class="text-[15px] font-medium">No incidents</div>
			<p class="text-subtle-foreground m-0 max-w-[400px] text-center text-[13px] leading-[1.55]">
				When something pages someone, it can be declared an incident and run from here, or
				straight from chat.
			</p>
			<Button variant="secondary" size="sm" onclick={() => (declaring = true)}>
				<SirenIcon data-icon="inline-start" />
				Declare an incident
			</Button>
		</div>
	{:else}
		<FilterBar
			services={data.filterOptions.services}
			teams={data.filterOptions.teams}
			people={data.filterOptions.leads}
		/>
		<div class="bg-card overflow-hidden rounded-xl border">
			<IncidentsTable incidents={data.incidents} filters={data.filters} now={data.now} />
		</div>
	{/if}
</Page>

<DeclareDialog bind:open={declaring} services={data.services} members={data.members} />
