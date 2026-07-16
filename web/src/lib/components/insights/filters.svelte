<script lang="ts">
	import DownloadIcon from '@lucide/svelte/icons/download';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { RANGE_OPTIONS, SERVICES, SEVERITIES, TEAMS, filterQuery, parseFilters, type Filters } from '$lib/insights';
	import { ws } from '$lib/navigation';

	let { tab }: { tab: string } = $props();

	// Controls render the sanitized scope, never raw URL params
	const filters = $derived(parseFilters(page.url));

	const DIMENSIONS = [
		{ key: 'team', placeholder: 'Team', width: 'w-[120px]', all: 'All teams', options: TEAMS },
		{ key: 'service', placeholder: 'Service', width: 'w-[140px]', all: 'All services', options: SERVICES },
		{ key: 'severity', placeholder: 'Severity', width: 'w-[110px]', all: 'All severities', options: SEVERITIES }
	] as const;

	function set(key: keyof Filters, value: string) {
		const query = filterQuery(filters, { [key]: value } as Partial<Filters>);
		goto(query || page.url.pathname, { keepFocus: true, noScroll: true, replaceState: true });
	}

	const exportHref = $derived.by(() => {
		const params = new URLSearchParams(filterQuery(filters));
		params.set('tab', tab);
		return ws(`/insights/export?${params}`);
	});
</script>

<div class="bg-card flex flex-wrap items-center gap-2.5 rounded-xl border px-3 py-2.5">
	{#each DIMENSIONS as dimension (dimension.key)}
		{@const active = filters[dimension.key]}
		<Select.Root
			type="single"
			value={active || 'all'}
			onValueChange={(value) => set(dimension.key, value === 'all' ? '' : value)}
		>
			<Select.Trigger
				size="sm"
				class={dimension.width}
				aria-label="Filter by {dimension.placeholder.toLowerCase()}"
			>
				{active || dimension.placeholder}
			</Select.Trigger>
			<Select.Content>
				<Select.Group>
					<Select.Item value="all" label={dimension.all}>{dimension.all}</Select.Item>
					{#each dimension.options as option (option)}
						<Select.Item value={option} label={option}>{option}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	{/each}

	<Select.Root type="single" value={filters.range} onValueChange={(value) => set('range', value)}>
		<Select.Trigger size="sm" class="w-[130px]" aria-label="Time range">
			{RANGE_OPTIONS.find((option) => option.value === filters.range)?.label}
		</Select.Trigger>
		<Select.Content>
			<Select.Group>
				{#each RANGE_OPTIONS as option (option.value)}
					<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
				{/each}
			</Select.Group>
		</Select.Content>
	</Select.Root>

	<div class="flex-1"></div>

	<Button variant="ghost" size="sm" href={exportHref} download="insights-{tab}.csv">
		<DownloadIcon data-icon="inline-start" />
		Export CSV
	</Button>
</div>
