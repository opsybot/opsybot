<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Tag from '$lib/components/tag.svelte';
	import * as InputGroup from '$lib/components/ui/input-group';
	import * as Select from '$lib/components/ui/select';
	import { SEVERITIES, STATUSES } from '$lib/alerts';

	let {
		sources,
		services,
		labels
	}: {
		sources: string[];
		services: string[];
		labels: string[];
	} = $props();

	const params = $derived(page.url.searchParams);
	const statuses = $derived(new Set(params.getAll('status').length ? params.getAll('status') : ['open', 'acked']));
	const range = $derived(params.get('range') ?? '24h');

	const RANGES = [
		{ value: '24h', label: 'Last 24 h' },
		{ value: '7d', label: 'Last 7 days' },
		{ value: '30d', label: 'Last 30 days' }
	];

	const FILTERS = $derived([
		{ key: 'severity', placeholder: 'Severity', width: 'w-[120px]', options: SEVERITIES as unknown as string[] },
		{ key: 'source', placeholder: 'Source', width: 'w-[130px]', options: sources },
		{ key: 'service', placeholder: 'Service', width: 'w-[140px]', options: services },
		{ key: 'label', placeholder: 'Label', width: 'w-[150px]', options: labels }
	]);

	const dirty = $derived(
		['q', 'severity', 'source', 'service', 'label'].some((key) => params.get(key))
	);

	function apply(next: URLSearchParams) {
		next.delete('cursor');
		goto(`?${next}`, { keepFocus: true, noScroll: true, replaceState: true });
	}

	function set(key: string, value: string | undefined) {
		const next = new URLSearchParams(params);
		if (value) next.set(key, value);
		else next.delete(key);
		apply(next);
	}

	function toggleStatus(status: string) {
		const next = new URLSearchParams(params);
		const current = new Set(statuses);
		if (current.has(status)) current.delete(status);
		else current.add(status);
		if (current.size === 0) return;

		next.delete('status');
		for (const value of current) next.append('status', value);
		apply(next);
	}

	function clear() {
		const next = new URLSearchParams();
		for (const status of statuses) next.append('status', status);
		next.set('range', range);
		apply(next);
	}
</script>

<div class="bg-card flex flex-wrap items-center gap-2.5 rounded-xl border px-3 py-2.5">
	<InputGroup.Root class="bg-inset border-border-strong h-[34px] w-[220px] rounded-md shadow-none">
		<InputGroup.Input
			placeholder="Search alerts"
			value={params.get('q') ?? ''}
			oninput={(event) => set('q', event.currentTarget.value)}
			class="text-[13px]"
		/>
		<InputGroup.Addon>
			<SearchIcon class="text-subtle-foreground size-4" />
		</InputGroup.Addon>
	</InputGroup.Root>

	<div class="flex gap-1.5">
		{#each STATUSES as status (status.id)}
			<Tag selected={statuses.has(status.id)} onclick={() => toggleStatus(status.id)}>
				{status.label}
			</Tag>
		{/each}
	</div>

	{#each FILTERS as filter (filter.key)}
		<Select.Root
			type="single"
			value={params.get(filter.key) ?? ''}
			onValueChange={(value) => set(filter.key, value || undefined)}
		>
			<Select.Trigger size="sm" class={filter.width}>
				{params.get(filter.key) ?? filter.placeholder}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each filter.options as option (option)}
						<Select.Item value={option} label={option}>{option}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	{/each}

	<Select.Root type="single" value={range} onValueChange={(value) => set('range', value)}>
		<Select.Trigger size="sm" class="w-[120px]">
			{RANGES.find((entry) => entry.value === range)?.label}
		</Select.Trigger>
		<Select.Content>
			<Select.Group>
				{#each RANGES as entry (entry.value)}
					<Select.Item value={entry.value} label={entry.label}>{entry.label}</Select.Item>
				{/each}
			</Select.Group>
		</Select.Content>
	</Select.Root>

	{#if dirty}
		<button
			type="button"
			onclick={clear}
			class="text-muted-foreground hover:text-brand-foreground text-[12.5px]"
		>
			Clear filters
		</button>
	{/if}
</div>
