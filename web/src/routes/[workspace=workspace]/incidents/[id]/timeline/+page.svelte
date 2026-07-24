<script lang="ts">
	import BracesIcon from '@lucide/svelte/icons/braces';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import { toast } from 'svelte-sonner';
	import Panel from '$lib/components/incidents/panel.svelte';
	import TimelineEntry from '$lib/components/incidents/timeline-entry.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Button } from '$lib/components/ui/button';
	import { ENTRY_TYPES, type EntryType } from '$lib/incidents';
	import { formatUtc } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const incident = $derived(data.incident);

	let filter = $state<EntryType | 'all'>('all');

	const visible = $derived(
		filter === 'all'
			? incident.timeline
			: incident.timeline.filter((entry) => entry.type === filter)
	);

	function exportTimeline(format: 'json' | 'markdown') {
		const payload =
			format === 'markdown'
				? `# ${incident.ref ?? incident.id}: timeline\n\n` +
					incident.timeline
						.map(
							(entry) =>
								`- **${formatUtc(entry.at)}** · ${entry.type} · ${entry.actor}: ${entry.text}`
						)
						.join('\n')
				: JSON.stringify(incident.timeline, null, 2);

		navigator.clipboard.writeText(payload);
		toast.success(
			`${format === 'markdown' ? 'Human-readable timeline' : 'Structured JSON'} copied to clipboard.`
		);
	}
</script>

<div class="flex max-w-[780px] flex-col gap-3.5">
	<div class="flex flex-wrap items-center gap-1.5">
		<Tag selected={filter === 'all'} onclick={() => (filter = 'all')}>All</Tag>
		{#each ENTRY_TYPES as entryType (entryType.id)}
			<Tag selected={filter === entryType.id} onclick={() => (filter = entryType.id)}>
				{entryType.label}
			</Tag>
		{/each}
		<div class="flex-1"></div>
		<Button variant="ghost" size="sm" onclick={() => exportTimeline('json')}>
			<BracesIcon data-icon="inline-start" />
			Export JSON
		</Button>
		<Button variant="ghost" size="sm" onclick={() => exportTimeline('markdown')}>
			<FileTextIcon data-icon="inline-start" />
			Export Markdown
		</Button>
	</div>

	<Panel class="px-4 pt-4 pb-1">
		{#if visible.length === 0}
			<p class="text-subtle-foreground m-0 mb-3.5 text-[13px]">No {filter} entries yet.</p>
		{:else}
			{#each visible as entry, index (entry.id)}
				<TimelineEntry {entry} last={index === visible.length - 1} />
			{/each}
		{/if}
	</Panel>
</div>
