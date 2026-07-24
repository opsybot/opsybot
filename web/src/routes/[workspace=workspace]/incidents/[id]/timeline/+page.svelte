<script lang="ts">
	import BracesIcon from '@lucide/svelte/icons/braces';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import { toast } from 'svelte-sonner';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Panel from '$lib/components/incidents/panel.svelte';
	import TimelineEntry from '$lib/components/incidents/timeline-entry.svelte';
	import Tag from '$lib/components/tag.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { ENTRY_TYPES, type EntryType } from '$lib/incidents';
	import { formatUtcDate } from '$lib/time';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const attachmentBase = $derived(
		`/${page.params.workspace}/incidents/${page.params.id}/attachments`
	);
	const revisions = $derived(form && 'revisions' in form ? form.revisions : []);
	const revisionEntryId = $derived(form && 'entryId' in form ? form.entryId : '');

	let tzOffset = $state('');
	$effect(() => {
		tzOffset = String(new Date().getTimezoneOffset());
	});

	let composing = $state(false);
	let backdated = $state(false);
	let category = $state<EntryType>('observation');
	let composeKey = $state('');

	function startComposing() {
		composeKey = crypto.randomUUID();
		composing = true;
	}

	function navigate(update: (params: URLSearchParams) => void) {
		const url = new URL(page.url);
		update(url.searchParams);
		goto(url, { keepFocus: true, noScroll: true });
	}

	function toggleFilter(type: EntryType) {
		const next = new Set(data.filter);
		if (next.has(type)) next.delete(type);
		else next.add(type);
		navigate((params) => {
			params.delete('type');
			for (const value of next) params.append('type', value);
			params.delete('cursor');
		});
	}

	function clearFilter() {
		navigate((params) => {
			params.delete('type');
			params.delete('cursor');
		});
	}

	function showPage(cursor: string | undefined) {
		navigate((params) => {
			if (cursor) params.set('cursor', cursor);
			else params.delete('cursor');
		});
	}

	$effect(() => {
		if (form && 'error' in form && form.error) toast.error(form.error);
	});
</script>

<div class="flex max-w-[780px] flex-col gap-3.5">
	<div class="flex flex-wrap items-center gap-1.5">
		<Tag selected={data.filter.length === 0} onclick={clearFilter}>All</Tag>
		{#each ENTRY_TYPES as entryType (entryType.id)}
			<Tag selected={data.filter.includes(entryType.id)} onclick={() => toggleFilter(entryType.id)}>
				{entryType.label}
			</Tag>
		{/each}
		<div class="flex-1"></div>
		<Button variant="ghost" size="sm" href="{page.url.pathname}/export">
			<BracesIcon data-icon="inline-start" />
			Export JSON
		</Button>
		<Button variant="ghost" size="sm" href="{page.url.pathname}/export?format=text">
			<FileTextIcon data-icon="inline-start" />
			Export text
		</Button>
	</div>

	<Panel class="px-4 pt-4 pb-1">
		{#if data.entries.length === 0}
			<p class="text-subtle-foreground m-0 mb-3.5 text-[13px]">
				{#if data.cursor}
					No further entries.
				{:else if data.filter.length > 0}
					No entries match the selected types.
				{:else}
					Nothing on the timeline yet.
				{/if}
			</p>
		{:else}
			{#each data.entries as entry, index (entry.id)}
				<TimelineEntry
					{entry}
					last={index === data.entries.length - 1}
					showDate={index === 0 ||
						formatUtcDate(entry.at) !== formatUtcDate(data.entries[index - 1].at)}
					{attachmentBase}
					revisions={revisionEntryId === entry.id ? revisions : []}
				/>
			{/each}
		{/if}
	</Panel>

	{#if data.nextCursor || data.cursor}
		<div class="flex items-center gap-2">
			{#if data.cursor}
				<Button variant="ghost" size="sm" onclick={() => showPage(undefined)}>
					Back to the start
				</Button>
			{/if}
			<div class="flex-1"></div>
			{#if data.nextCursor}
				<Button variant="outline" size="sm" onclick={() => showPage(data.nextCursor)}>
					Later entries →
				</Button>
			{/if}
		</div>
	{/if}

	<Panel class="p-4">
		{#if composing}
			<form
				method="POST"
				action="?/addEntry"
				class="flex flex-col gap-2.5"
				use:enhance={() =>
					async ({ result, update }) => {
						if (result.type === 'success') {
							composing = false;
							backdated = false;
							if (data.nextCursor) toast.info('Entry added at the end of the timeline.');
						}
						await update({ reset: true });
					}}
			>
				<input type="hidden" name="idempotencyKey" value={composeKey} />
				<Textarea
					name="text"
					rows={3}
					placeholder="What happened? Describe the event, not who is at fault."
					aria-label="Timeline entry"
				/>
				<div class="flex flex-wrap items-center gap-1.5">
					{#each ENTRY_TYPES as entryType (entryType.id)}
						<label
							class="inline-flex h-6 cursor-pointer items-center rounded-md border px-2.5 text-xs font-medium has-[:focus-visible]:border-brand-edge has-[:focus-visible]:ring-brand-edge/50 has-[:focus-visible]:ring-2 {category ===
							entryType.id
								? 'bg-brand-wash border-brand-edge text-brand-foreground'
								: 'bg-popover border-input text-muted-foreground'}"
						>
							<input
								type="radio"
								name="category"
								value={entryType.id}
								class="sr-only"
								bind:group={category}
							/>
							{entryType.label}
						</label>
					{/each}
				</div>
				<label class="text-muted-foreground flex items-center gap-2 text-[12.5px]">
					<input type="checkbox" bind:checked={backdated} />
					This happened earlier
				</label>
				{#if backdated}
					<div class="flex flex-col gap-1">
						<input type="hidden" name="tzOffset" value={tzOffset} />
						<Input
							name="at"
							type="datetime-local"
							aria-label="When it happened, in your local time"
						/>
						<span class="text-subtle-foreground text-[11.5px]">
							Entered in your local time, stored and shown in UTC. Marked retroactive.
						</span>
					</div>
				{/if}
				<div class="flex items-center gap-2">
					<div class="flex-1"></div>
					<Button variant="ghost" size="sm" onclick={() => (composing = false)}>Cancel</Button>
					<Button type="submit" size="sm">Add entry</Button>
				</div>
			</form>
		{:else}
			<Button variant="outline" size="sm" onclick={startComposing}>
				<PlusIcon data-icon="inline-start" />
				Add timeline entry
			</Button>
		{/if}
	</Panel>
</div>
