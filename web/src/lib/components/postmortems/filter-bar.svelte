<script lang="ts">
	import SearchIcon from '@lucide/svelte/icons/search';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { SERVICES } from '$lib/incidents';

	let {
		query,
		service,
		severity,
		range
	}: {
		query: string;
		service: string;
		severity: string;
		range: string;
	} = $props();

	const RANGES = [
		{ value: '30d', label: 'Last 30 days' },
		{ value: '90d', label: 'Last quarter' },
		{ value: 'all', label: 'All time' }
	];

	const SEVERITIES = ['SEV1', 'SEV2', 'SEV3'];

	function set(key: string, value: string) {
		const url = new URL(page.url);
		if (value) url.searchParams.set(key, value);
		else url.searchParams.delete(key);
		goto(url, { keepFocus: true, noScroll: true, replaceState: true });
	}

	const filtered = $derived(!!(query || service || severity));

	let typed: ReturnType<typeof setTimeout>;
</script>

<div class="bg-card flex flex-wrap items-center gap-2.5 rounded-xl border px-3 py-2.5">
	<div class="relative w-[210px]">
		<SearchIcon
			class="text-subtle-foreground pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2"
		/>
		<Input
			value={query}
			placeholder="Search postmortems"
			aria-label="Search postmortems"
			class="h-[34px] pl-[30px] text-[13px]"
			oninput={(event: Event) => {
				const { value } = event.currentTarget as HTMLInputElement;
				clearTimeout(typed);
				typed = setTimeout(() => set('q', value), 200);
			}}
		/>
	</div>

	<Select.Root type="single" value={service} onValueChange={(value) => set('service', value)}>
		<Select.Trigger size="sm" class="w-[140px]">{service || 'Service'}</Select.Trigger>
		<Select.Content>
			<Select.Group>
				{#each SERVICES as name (name)}
					<Select.Item value={name} label={name}>{name}</Select.Item>
				{/each}
			</Select.Group>
		</Select.Content>
	</Select.Root>

	<Select.Root type="single" value={severity} onValueChange={(value) => set('severity', value)}>
		<Select.Trigger size="sm" class="w-[110px]">{severity || 'Severity'}</Select.Trigger>
		<Select.Content>
			<Select.Group>
				{#each SEVERITIES as level (level)}
					<Select.Item value={level} label={level}>{level}</Select.Item>
				{/each}
			</Select.Group>
		</Select.Content>
	</Select.Root>

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

	{#if filtered}
		<a
			href="/postmortems"
			class="text-muted-foreground hover:text-brand-foreground text-[12.5px]"
		>
			Clear
		</a>
	{/if}
</div>
