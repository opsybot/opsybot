<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Tag from '$lib/components/tag.svelte';
	import * as InputGroup from '$lib/components/ui/input-group';
	import * as Select from '$lib/components/ui/select';
	import { SEVERITIES, STAGES } from '$lib/incidents';

	let {
		services = [],
		teams = [],
		people = []
	}: { services?: string[]; teams?: string[]; people?: string[] } = $props();

	const params = $derived(page.url.searchParams);
	const preset = $derived(params.get('preset') ?? 'active');
	const range = $derived(params.get('range') ?? '30d');

	const FILTERS = $derived([
		{ key: 'severity', placeholder: 'Severity', width: 'w-[110px]', options: SEVERITIES.map((s) => s.id) },
		{ key: 'status', placeholder: 'Status', width: 'w-[130px]', options: STAGES as unknown as string[] },
		{ key: 'service', placeholder: 'Service', width: 'w-[140px]', options: services },
		{ key: 'team', placeholder: 'Team', width: 'w-[120px]', options: teams },
		{ key: 'lead', placeholder: 'Lead', width: 'w-[130px]', options: people }
	]);

	const RANGES = [
		{ value: '7d', label: 'Last 7 days' },
		{ value: '30d', label: 'Last 30 days' },
		{ value: '90d', label: 'Last quarter' }
	];

	const dirty = $derived(
		['q', 'severity', 'status', 'service', 'team', 'lead'].some((key) => params.get(key))
	);

	function set(key: string, value: string | undefined) {
		const next = new URLSearchParams(params);
		if (value) next.set(key, value);
		else next.delete(key);
		goto(`?${next}`, { keepFocus: true, noScroll: true, replaceState: true });
	}

	function clear() {
		const next = new URLSearchParams();
		next.set('preset', preset);
		next.set('range', range);
		goto(`?${next}`, { keepFocus: true, noScroll: true, replaceState: true });
	}
</script>

<div class="bg-card flex flex-wrap items-center gap-2.5 rounded-xl border px-3 py-2.5">
	<div class="flex gap-1.5">
		{#each [['active', 'Active'], ['all', 'All'], ['mine', 'Mine']] as [value, label] (value)}
			<Tag selected={preset === value} onclick={() => set('preset', value)}>{label}</Tag>
		{/each}
	</div>

	<span class="bg-border h-5 w-px"></span>

	<InputGroup.Root class="bg-inset border-border-strong h-[34px] w-[190px] rounded-md shadow-none">
		<InputGroup.Input
			placeholder="Search incidents"
			value={params.get('q') ?? ''}
			oninput={(event) => set('q', event.currentTarget.value)}
			class="text-[13px]"
		/>
		<InputGroup.Addon>
			<SearchIcon class="text-subtle-foreground size-4" />
		</InputGroup.Addon>
	</InputGroup.Root>

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
		<Select.Trigger size="sm" class="w-[130px]">
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
			Clear
		</button>
	{/if}
</div>
