<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as InputGroup from '$lib/components/ui/input-group';
	import * as Select from '$lib/components/ui/select';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { isFailure } from '$lib/admin';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const ACTION_TYPES = [
		['auth', 'Authentication'],
		['member', 'Members'],
		['team', 'Teams'],
		['key', 'API keys'],
		['sso', 'SSO'],
		['channel', 'Channels'],
		['workspace', 'Workspace']
	] as const;
	const ACTION_LABELS: Record<string, string> = Object.fromEntries(ACTION_TYPES);

	const params = $derived(page.url.searchParams);
	const filter = $derived({
		q: params.get('q') ?? '',
		actor: params.get('actor') ?? '',
		action: params.get('action') ?? '',
		cursor: params.get('cursor') ?? ''
	});
	const dirty = $derived(!!(filter.q || filter.actor || filter.action));
	const actorName = $derived(data.members.find((member) => member.id === filter.actor)?.name ?? '');

	function set(key: string, value: string | undefined) {
		const next = new URLSearchParams(params);
		if (value) next.set(key, value);
		else next.delete(key);
		if (key !== 'cursor') next.delete('cursor');
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
			<Select.Trigger size="sm" class="w-[170px]" aria-label="Actor">{actorName || 'Any actor'}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each data.members as member (member.id)}
						<Select.Item value={member.id} label={member.name}>{member.name}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
		<Select.Root type="single" value={filter.action} onValueChange={(value) => set('action', value || undefined)}>
			<Select.Trigger size="sm" class="w-[160px]" aria-label="Action type">
				{ACTION_LABELS[filter.action] ?? 'Any action'}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each ACTION_TYPES as [value, label] (value)}
						<Select.Item {value} {label}>{label}</Select.Item>
					{/each}
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
				{#each data.entries as entry (entry.id)}
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
		{#if data.nextCursor || filter.cursor}
			<div class="flex items-center gap-2 border-t px-[18px] py-2.5">
				{#if filter.cursor}
					<Button size="sm" variant="ghost" onclick={() => set('cursor', undefined)}>Back to latest</Button>
				{/if}
				<div class="flex-1"></div>
				{#if data.nextCursor}
					<Button size="sm" variant="secondary" onclick={() => set('cursor', data.nextCursor)}>Older →</Button>
				{/if}
			</div>
		{/if}
	</section>
</div>
