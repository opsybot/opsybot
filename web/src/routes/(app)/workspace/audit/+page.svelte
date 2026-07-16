<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as InputGroup from '$lib/components/ui/input-group';
	import * as Select from '$lib/components/ui/select';
	import * as Table from '$lib/components/ui/table';
	import { auditActionPrefix, auditMatches, isFailure } from '$lib/admin';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const params = $derived(page.url.searchParams);
	const filter = $derived({
		q: params.get('q') ?? '',
		actor: params.get('actor') ?? '',
		action: params.get('action') ?? ''
	});
	const dirty = $derived(!!(filter.q || filter.actor || filter.action));

	const actors = $derived([...new Set(data.entries.map((entry) => entry.actor))]);
	const actions = $derived([...new Set(data.entries.map((entry) => auditActionPrefix(entry.action)))]);
	const visible = $derived(data.entries.filter((entry) => auditMatches(entry, filter)));

	function set(key: string, value: string | undefined) {
		const next = new URLSearchParams(params);
		if (value) next.set(key, value);
		else next.delete(key);
		goto(`?${next}`, { keepFocus: true, noScroll: true, replaceState: true });
	}
	function clear() {
		goto(page.url.pathname, { keepFocus: true, noScroll: true, replaceState: true });
	}
</script>

<div class="flex max-w-[920px] flex-col gap-3.5">
	<div class="bg-card flex flex-wrap items-center gap-2.5 rounded-xl border px-3 py-2.5">
		<InputGroup.Root class="bg-inset border-border-strong h-[34px] w-[230px] rounded-md shadow-none">
			<InputGroup.Input
				placeholder="Search actions and targets"
				value={filter.q}
				oninput={(event) => set('q', event.currentTarget.value)}
				class="text-[13px]"
			/>
			<InputGroup.Addon>
				<SearchIcon class="text-subtle-foreground size-4" />
			</InputGroup.Addon>
		</InputGroup.Root>
		<Select.Root type="single" value={filter.actor} onValueChange={(value) => set('actor', value || undefined)}>
			<Select.Trigger size="sm" class="w-[160px]" aria-label="Actor">{filter.actor || 'Actor'}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each actors as actor (actor)}<Select.Item value={actor} label={actor}>{actor}</Select.Item>{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
		<Select.Root type="single" value={filter.action} onValueChange={(value) => set('action', value || undefined)}>
			<Select.Trigger size="sm" class="w-[150px]" aria-label="Action type">{filter.action || 'Action type'}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each actions as action (action)}<Select.Item value={action} label={action}>{action}</Select.Item>{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
		{#if dirty}
			<button type="button" onclick={clear} class="text-muted-foreground hover:text-brand-foreground text-[12.5px]">
				Clear
			</button>
		{/if}
		<div class="flex-1"></div>
		<span class="text-subtle-foreground text-[12px]">export & streaming live in compliance settings</span>
	</div>

	<section class="bg-card overflow-hidden rounded-xl border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head class="pl-[18px]">Time</Table.Head>
					<Table.Head>Actor</Table.Head>
					<Table.Head>Action</Table.Head>
					<Table.Head>Target</Table.Head>
					<Table.Head class="pr-[18px]">Source IP</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each visible as entry (entry.id)}
					<Table.Row data-audit={entry.id}>
						<Table.Cell class="text-muted-foreground pl-[18px] font-mono text-[12px] whitespace-nowrap">{entry.at}</Table.Cell>
						<Table.Cell class="text-[12.5px] font-medium">{entry.actor}</Table.Cell>
						<Table.Cell>
							<span class="font-mono text-[12px] {isFailure(entry.action) ? 'text-critical-ink' : 'text-foreground'}">
								{entry.action}
							</span>
						</Table.Cell>
						<Table.Cell class="text-muted-foreground text-[12.5px]">{entry.target}</Table.Cell>
						<Table.Cell class="text-muted-foreground pr-[18px] font-mono text-[12px]">{entry.ip}</Table.Cell>
					</Table.Row>
				{:else}
					<Table.Row>
						<Table.Cell colspan={5} class="text-subtle-foreground py-8 text-center text-[13px]">
							No matching entries.
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</section>
</div>
