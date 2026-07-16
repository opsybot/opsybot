<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import WrenchIcon from '@lucide/svelte/icons/wrench';
	import Page from '$lib/components/layout/page.svelte';
	import MaintenanceDialog from '$lib/components/statuspages/maintenance-dialog.svelte';
	import SpTabs from '$lib/components/statuspages/sp-tabs.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { formatUtc } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let scheduling = $state(false);
</script>

{#snippet window(entry: (typeof data.upcoming)[number], past: boolean)}
	<div class="flex items-center gap-3 border-t px-4 py-3 first:border-t-0" class:opacity-70={past}>
		<span
			class="bg-inset text-subtle-foreground flex size-8 shrink-0 items-center justify-center rounded-sm border"
		>
			<WrenchIcon class="size-[14px]" />
		</span>
		<div class="min-w-0 flex-1">
			<div class="text-[13.5px] font-medium">{entry.title}</div>
			<div class="text-subtle-foreground mt-[3px] font-mono text-[11px]">
				{formatUtc(entry.startsAt)} → {formatUtc(entry.endsAt)} · notice {entry.notice}
			</div>
		</div>
		<div class="flex flex-wrap justify-end gap-[5px]">
			{#each entry.components as component (component)}
				<Tag>{component}</Tag>
			{/each}
		</div>
	</div>
{/snippet}

<Page title="Status pages" subtitle="Tell customers before they ask">
	<SpTabs current="maintenance" />

	<div class="mt-3.5 flex max-w-[760px] flex-col gap-3.5">
		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="flex items-center gap-2.5 border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Upcoming</span>
				<Badge tone="neutral" size="sm">{data.upcoming.length}</Badge>
				<div class="flex-1"></div>
				<Button size="sm" onclick={() => (scheduling = true)}>
					<PlusIcon data-icon="inline-start" />
					Schedule maintenance
				</Button>
			</header>

			{#each data.upcoming as entry (entry.id)}
				{@render window(entry, false)}
			{:else}
				<p class="text-subtle-foreground m-0 px-4 py-5 text-[12.5px]">Nothing scheduled.</p>
			{/each}
		</section>

		<section class="bg-card overflow-hidden rounded-xl border">
			<header class="border-b px-4 py-3">
				<span class="text-[13.5px] font-semibold">Past</span>
			</header>

			{#each data.past as entry (entry.id)}
				{@render window(entry, true)}
			{:else}
				<p class="text-subtle-foreground m-0 px-4 py-5 text-[12.5px]">No past maintenance.</p>
			{/each}
		</section>
	</div>
</Page>

<MaintenanceDialog bind:open={scheduling} components={data.components} notices={data.notices} />
