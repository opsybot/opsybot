<script lang="ts">
	import BracesIcon from '@lucide/svelte/icons/braces';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { enhance } from '$app/forms';
	import { toast } from 'svelte-sonner';
	import Panel from '$lib/components/incidents/panel.svelte';
	import TimelineEntry from '$lib/components/incidents/timeline-entry.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import * as Field from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Textarea } from '$lib/components/ui/textarea';
	import { ENTRY_TYPES, type EntryType } from '$lib/incidents';
	import { formatUtc } from '$lib/time';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const incident = $derived(data.incident);

	let filter = $state<EntryType | 'all'>('all');
	let type = $state<EntryType>('observation');
	let text = $state('');
	let retroText = $state('');
	let retroTime = $state('09:25');

	const visible = $derived(
		filter === 'all'
			? incident.timeline
			: incident.timeline.filter((entry) => entry.type === filter)
	);

	function exportTimeline(format: 'json' | 'markdown') {
		const payload =
			format === 'markdown'
				? `# ${incident.id}: timeline\n\n` +
					incident.timeline
						.map(
							(entry) =>
								`- **${formatUtc(entry.at)}** · ${entry.type} · ${entry.actor}${entry.retro ? ' _(retroactive)_' : ''}: ${entry.text}`
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

	<Panel class="flex flex-col gap-2.5 p-3.5">
		<form
			method="POST"
			action="?/entry"
			use:enhance={() => async ({ update }) => {
				await update();
				text = '';
			}}
			class="flex flex-col gap-2.5"
		>
			<div class="flex items-center gap-2">
				<Select.Root type="single" name="type" bind:value={type}>
					<Select.Trigger size="sm" class="w-[150px]">
						{ENTRY_TYPES.find((entry) => entry.id === type)?.label}
					</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each ENTRY_TYPES as entryType (entryType.id)}
								<Select.Item value={entryType.id} label={entryType.label}>
									{entryType.label}
								</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>
				<span class="text-subtle-foreground text-[11.5px]">timestamped now, in UTC</span>
			</div>

			<Textarea name="text" rows={2} bind:value={text} placeholder="What happened?" />

			<div class="flex justify-end">
				<Button type="submit" size="sm" disabled={!text.trim()}>
					<PlusIcon data-icon="inline-start" />
					Add entry
				</Button>
			</div>
		</form>
	</Panel>

	<Panel>
		<Collapsible.Root>
			<Collapsible.Trigger
				class="text-muted-foreground hover:text-foreground m-3.5 inline-flex items-center gap-2 rounded-md border px-2.5 py-[7px] text-[13px] font-medium"
			>
				<HistoryIcon class="size-[15px]" />
				Add a retroactive entry
			</Collapsible.Trigger>

			<Collapsible.Content>
				<form
					method="POST"
					action="?/entry"
					use:enhance={() => async ({ update }) => {
						await update();
						retroText = '';
					}}
					class="flex flex-col gap-2.5 px-3.5 pb-3.5"
				>
					<input type="hidden" name="type" value={type} />
					<p class="text-subtle-foreground m-0 text-[12.5px] leading-[1.55]">
						For things that happened but weren't captured live. The entry sorts by its real time and
						is visibly marked retroactive.
					</p>

					<div class="flex flex-wrap items-start gap-2">
						<Field.Field class="w-[140px] gap-1.5 space-y-0">
							<Field.FieldLabel for="at" class="text-muted-foreground text-[13px] font-medium">
								Actual time (UTC)
							</Field.FieldLabel>
							<Input id="at" name="at" type="time" bind:value={retroTime} class="h-[34px] text-[13px]" />
						</Field.Field>

						<Field.Field class="min-w-[220px] flex-1 gap-1.5 space-y-0">
							<Field.FieldLabel for="retro" class="text-muted-foreground text-[13px] font-medium">
								What happened
							</Field.FieldLabel>
							<Input
								id="retro"
								name="text"
								bind:value={retroText}
								placeholder="Paged DBA on-call out of band"
								class="h-[34px] text-[13px]"
							/>
						</Field.Field>
					</div>

					<Button type="submit" size="sm" variant="secondary" class="self-start" disabled={!retroText.trim()}>
						Add retroactive entry
					</Button>
				</form>
			</Collapsible.Content>
		</Collapsible.Root>
	</Panel>
</div>
